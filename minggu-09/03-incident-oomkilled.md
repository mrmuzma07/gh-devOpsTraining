# Minggu 9 — Modul 03: Incident OOMKilled

> **"Aplikasi jalan sebentar, terus tiba-tiba mati dan restart."**

OOMKilled = **Out Of Memory Killed**. Pod mati bukan karena bug code, tapi karena proses di dalam container **memakan RAM lebih banyak dari yang diizinkan (memory limit)**. Kernel Linux mengirim sinyal **SIGKILL (signal 9)** untuk menyelamatkan node dari kehabisan memory, dan exit code yang tercatat adalah **137** (128 + 9).

Analogi: seperti kamu menyewa kamar kos dengan容量 10 m², lalu kamu bawa 15 m³ barang. Penjaga kos akan强制 mengusir barang lebih. Mirip dengan itu, kubelet akan强制 mematikan proses yang pakai RAM melebihi limit.

---

## 🎯 Tujuan Modul

1. Membedakan **OOMKilled** (exit 137) dari bug code (exit 1)
2. Membaca **memory metrics** dari Mimir untuk konfirmasi
3. Menentukan apakah perlu **naikkan limit** atau **fix memory leak**
4. Menerapkan **memory requests + limits** yang realistis
5. Membangun **alert** berbasis memory usage (kamu sudah punya di Minggu 8!)

---

## 🔑 Konsep Kunci: Memory Requests vs Limits

Sebelum masuk incident, pahami dulu konsep ini karena **sangat kritis**:

```
┌──────────────────────────────────────────────────────────┐
│                MEMORY MANAGEMENT K8s                     │
├──────────────────────────────────────────────────────────┤
│                                                          │
│   Requests (128Mi)    Limits (256Mi)                     │
│   ◄──────────►        ◄──────────►                       │
│   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓     ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                  │
│   ░░░░░░░░░░░░░░░░     ░░░░░░░░░░░░░░░░                  │
│   Garansi minimum     Batas atas (hard cap)              │
│   Bisa diisi sampai   Jika lewat → OOMKilled!            │
│   limit               ▓▓▓░░ = ~200Mi aktual              │
│                                                          │
│   ❌ Tanpa requests: pod bisa tidak dijadwalkan          │
│   ❌ Tanpa limits: pod bisa makan semua RAM node         │
│   ✅ Best practice: requests < real usage < limits       │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**Memory pressure di node** — saat node kehabisan RAM, K8s bisa memilih pod untuk **evict** (bukan kill, tapi matikan dengan halus). Eviction terjadi ketika:

- Node memory usage > threshold (default 100Mi free)
- Pod memory usage > request (pod yang tidak punya request atau pakai memory > request duluan)

---

## 📦 Simulasi: Reproduce OOMKilled

Kita akan deploy aplikasi yang **sengaja makan RAM banyak**.

### File: `manifests/02-oomkilled-stress.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oom-app
  namespace: insiden-lab
  labels:
    app: oom-app
    lab: insiden-2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: oom-app
  template:
    metadata:
      labels:
        app: oom-app
    spec:
      containers:
      - name: stress
        image: polinux/stress:1.2.3
        args:
        # Alokasi 200MB virtual memory → melewati limit 128Mi
        - --vm
        - "1"
        - --vm-bytes
        - "200M"
        - --timeout
        - "600s"
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            # Sengaja lebih kecil dari vm-bytes
            cpu: 200m
            memory: 128Mi
```

### Deploy & Observe

```bash
kubectl apply -f manifests/02-oomkilled-stress.yaml

# Watch pod (akan restart beberapa kali)
kubectl get pods -n insiden-lab -l app=oom-app -w
```

**Output yang diharapkan (dalam 1-2 menit):**
```
NAME                       READY   STATUS      RESTARTS   AGE
oom-app-6b8d9f7c8d-abcd1   1/1     Running     0          5s
oom-app-6b8d9f7c8d-abcd1   0/1     OOMKilled   1          45s
oom-app-6b8d9f7c8d-abcd1   1/1     Running     0          50s
oom-app-6b8d9f7c8d-abcd1   0/1     OOMKilled   2          95s
```

---

## 🔍 Symptoms

### Symptom 1 — Status menunjukkan `OOMKilled` di Last State

```bash
$ kubectl get pods -n insiden-lab -l app=oom-app
NAME                       READY   STATUS      RESTARTS   AGE
oom-app-6b8d9f7c8d-abcd1   0/1     OOMKilled   3          5m
```

**Penting:** `STATUS=OOMKilled` berarti **Last State** adalah OOMKilled, BUKAN current state. Current state sebenarnya adalah `Waiting: CrashLoopBackOff` — tapi K8s menampilkan "OOMKilled" di kolom STATUS untuk强调 bahwa root cause-nya adalah OOM.

### Symptom 2 — Memory usage tinggi di dashboard

Buka dashboard **"Pod Resource Usage"** dari Minggu 8:
- Panel "Memory Usage by Pod" → grafik naik terus sampai limit
- Panel "OOM Events" → counter naik

### Symptom 3 — Alert firing (yang kamu setup di Minggu 8)

```
[FIRING] PodOOMKilled
  namespace = insiden-lab
  pod       = oom-app-6b8d9f7c8d-abcd1
  container = stress
  reason    = Error
  message   = container killed due to memory usage
  severity  = critical
```

---

## 🕵️ Investigation

### Step 1 — describe pod (cari exit code)

```bash
$ kubectl describe pod oom-app-6b8d9f7c8d-abcd1 -n insiden-lab
```

Bagian penting di output:

```
Containers:
  stress:
    ...
    State:          Waiting
      Reason:       CrashLoopBackOff
    Last State:     Terminated
      Reason:       OOMKilled       ← ★ INI KUNCINYA ★
      Exit Code:    137             ← 128 + 9 (SIGKILL)
      Started:      Mon, 10 Aug 2026 09:45:23 +0700
      Finished:     Mon, 10 Aug 2026 09:46:01 +0700  ← 38 detik!
    Ready:          False
    Restart Count:  3
    Limits:
      cpu:     200m
      memory:  128Mi               ← INI BATAS YANG DILANGGAR

Events:
  Type     Reason     Age   From      Message
  ----     ------     ----  ----      -------
  Warning  BackOff    2m    kubelet   Back-off restarting failed container stress
```

**Tiga hal yang harus kamu catat:**
1. **`Reason: OOMKilled`** — bukan "Error" (yang artinya bug)
2. **`Exit Code: 137`** — spesifik memory kill
3. **`Limits.memory: 128Mi`** — batas yang dilanggar

### Step 2 — Cek memory metrics real-time

```bash
$ kubectl top pod oom-app-6b8d9f7c8d-abcd1 -n insiden-lab
NAME                       CPU(cores)   MEMORY(bytes)
oom-app-6b8d9f7c8d-abcd1   45m          128Mi        ← MEETING LIMIT
```

Atau dari Mimir via Grafana Explore:

```promql
# Memory usage pod dalam bytes
container_memory_working_set_bytes{
  namespace="insiden-lab",
  pod="oom-app-6b8d9f7c8d-abcd1",
  container="stress"
}

# Memory limit
kube_pod_container_resource_limits{
  namespace="insiden-lab",
  resource="memory"
}

# Ratio usage / limit
container_memory_working_set_bytes / 
kube_pod_container_resource_limits{resource="memory"}
```

### Step 3 — Cek apakah node juga OOM (system-wide)

```bash
$ kubectl describe node laptop-k3s | grep -A 5 "Memory"
```

```
MemoryPressure   False        ← node tidak kehabisan memory
Allocatable:
  memory:  8Gi
Allocated:
  memory:  2.3Gi (29%)
```

**Kesimpulan:** Node sehat. Yang OOM hanya container `stress` karena limitnya sendiri yang terlampaui.

### Step 4 — Cross-check Loki untuk konteks

```logql
{namespace="insiden-lab"} |= "OOM" |= "memory"
```

---

## 🎯 Root Cause Analysis

### 5 Whys

```
Problem: Pod oom-app OOMKilled terus

  Why 1: Kenapa OOMKilled?
    → Memory usage > limit (128Mi)
  
  Why 2: Kenapa memory usage > 128Mi?
    → App sengaja alokasi 200MB (vm-bytes=200M)
  
  Why 3: Kenapa alokasi 200MB tapi limit cuma 128Mi?
    → Limit di-set terlalu kecil untuk workload
  
  Why 4: Kenapa limit di-set terlalu kecil?
    → Copy-paste dari template tanpa profiling real usage
  
  Why 5: Kenapa tidak ada profiling?
    → Tidak ada baseline test di staging → prod
       tidak bisa prediksi memory usage

🎯 ROOT CAUSE:
   Limit memory terlalu rendah untuk workload aktual
```

### 3 Skenario Root Cause OOMKilled

| # | Skenario | Ciri-ciri |
|---|---|---|
| 1 | **Limit terlalu rendah** | Steady growth sampai limit, baru OOM. **Solusi:** naikkan limit. |
| 2 | **Memory leak di code** | Usage naik terus tanpa henti sampai OOM, setelah restart naik lagi ke titik tinggi. **Solusi:** fix bug di code. |
| 3 | **Burst workload** | Usage spike mendadak (misal: import CSV besar) lalu turun. **Solusi:** tambah JVM heap tuning atau streaming processing. |

**Cara bedakan di grafik Mimir:**
- **Skenario 1:** Grafik landai, naik gradual, plateau di limit
- **Skenario 2:** Grafik naik tajam dan tidak pernah turun (setiap restart mulai tinggi)
- **Skenario 3:** Grafik ada spike tajam lalu turun

---

## 🛠️ Mitigation

### Opsi A — Naikkan Limit (kalau skenario 1)

```bash
kubectl set resources deployment/oom-app -n insiden-lab   --limits=memory=512Mi   --requests=memory=256Mi

# Verify
kubectl get pod -n insiden-lab -l app=oom-app -o jsonpath='{.items[0].spec.containers[0].resources}'
# Output: {"limits":{"cpu":"200m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"256Mi"}}
```

### Opsi B — Scale untuk Distribusi Beban

```bash
# Tambah replicas supaya total memory lebih besar
kubectl scale deployment/oom-app -n insiden-lab --replicas=3
# Total memory: 3 × 128Mi = 384Mi (tapi masing-masing masih 128Mi)
```

**WARNING:** Scale bukan solusi untuk memory leak! Kalau ada leak, tiap pod akan tetap OOM.

### Opsi C — Restart & Investigasi Code (skenario 2)

```bash
# Restart sekarang untuk redam dampak
kubectl rollout restart deployment/oom-app -n insiden-lab

# Capture heap dump SEBELUM restart untuk dianalisis
# (Untuk Java/Go app, perlu tool khusus)
kubectl exec -n insiden-lab <pod-name> -- jmap -dump:live,format=b,file=/tmp/heap.bin 1

# Copy heap dump ke lokal untuk dianalisis dengan Eclipse MAT / pprof
kubectl cp insiden-lab/<pod-name>:/tmp/heap.bin ./heap.bin
```

### Opsi D — Kill pod sekarang (jika benar-benar stuck)

```bash
# Darurat: paksa matiin pod
kubectl delete pod -n insiden-lab -l app=oom-app
# ReplicaSet akan bikin pod baru dengan memory fresh
```

---

## 🛡️ Prevention

### Prevention 1 — Set Realistic Memory Requests + Limits

```yaml
spec:
  containers:
  - name: go-app
    resources:
      # Requests = rata-rata usage (untuk scheduler)
      requests:
        memory: 256Mi
        cpu: 100m
      # Limits = peak usage + buffer 30-50%
      limits:
        memory: 512Mi
        cpu: 500m
```

**Cara menghitung:**
```
Real avg usage dari staging (misal 200Mi)
Peak usage + 30% buffer = 260Mi → set limit 256Mi
```

### Prevention 2 — Pakai HorizontalPodAutoscaler (HPA) untuk Distribusi

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: oom-app-hpa
  namespace: insiden-lab
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: oom-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 70  # scale jika > 70% memory
```

### Prevention 3 — Memory Profiling di CI/CD

Tambahkan step di GitLab CI:

```yaml
memory-profile:
  stage: test
  image: golang:1.22
  script:
  - go test -memprofile=mem.prof ./...
  - go tool pprof -top -cum mem.prof > memory-report.txt
  artifacts:
    paths:
    - memory-report.txt
  allow_failure: false
```

### Prevention 4 — JVM Heap Tuning (untuk Java app)

```yaml
env:
- name: JAVA_OPTS
  value: >-
    -XX:+UseContainerSupport
    -XX:MaxRAMPercentage=75.0
    -XX:+HeapDumpOnOutOfMemoryError
    -XX:HeapDumpPath=/tmp/heapdump.hprof
```

**Penjelasan:**
- `UseContainerSupport` → JVM hormati cgroup limit K8s
- `MaxRAMPercentage=75` → JVM heap max hanya 75% dari container memory (sisa untuk non-heap)
- `HeapDumpOnOutOfMemoryError` → otomatis dump heap saat OOM untuk debugging

### Prevention 5 — Quality of Service (QoS) Guarantee

Pastikan pod masuk kategori **Guaranteed** atau **Burstable**:

| QoS Class | Cara Dapat | Perilaku Saat Pressure |
|---|---|---|
| **Guaranteed** | requests == limits untuk semua container | **Tidak pernah di-evict** pertama kali |
| **Burstable** | requests < limits (minimal salah satu container) | Di-evict setelah BestEffort |
| **BestEffort** | Tidak ada requests/limits | Di-evict pertama kali |

**Rekomendasi:** Untuk aplikasi production, targetkan **Burstable** atau **Guaranteed**. Jangan pernah BestEffort.

---

## 🧪 Verifikasi Recovery

```bash
# 1. Pod Running stabil
kubectl get pods -n insiden-lab -l app=oom-app -w
# Harus stabil di 1/1 Running, RESTARTS tidak naik

# 2. Memory usage di bawah limit
kubectl top pod -n insiden-lab -l app=oom-app
NAME                       CPU(cores)   MEMORY(bytes)
oom-app-6b8d9f7c8d-abcd1   45m          200Mi        ← di bawah 512Mi

# 3. Alert resolved
# Cek Slack: "[RESOLVED] PodOOMKilled" atau di Grafana Alerting

# 4. Stress test konfirmasi limit cukup
kubectl exec -n insiden-lab <pod-name> --   stress --vm 1 --vm-bytes 300M --timeout 30s
# Harus selesai tanpa OOM
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/02-oomkilled-stress.yaml
```

---

## 📖 Rangkuman

| Aspek | Catatan |
|---|---|
| **Tanda OOMKilled** | Exit code 137, Reason OOMKilled, Last State Terminated |
| **Root cause umum** | Limit terlalu rendah, memory leak, burst workload |
| **Mitigation cepat** | Naikkan limit, scale, restart, delete pod |
| **Cara deteksi memory leak** | Grafik naik tiap restart, atau heap dump analysis |
| **Prevention** | Realistic limits, HPA, JVM tuning, QoS class |
| **Alert yang firing** | `PodOOMKilled` atau `ContainerMemoryHigh` (Minggu 8) ✅ |

---

## ➡️ Modul 04: Pending Pod

Sekarang kamu sudah paham container-level incident (CrashLoop, OOM). Modul selanjutnya membahas **scheduling-level incident**: pod yang Pending karena tidak bisa dijadwalkan ke node mana pun.

👉 Lanjut ke `04-incident-pending-pod.md`
