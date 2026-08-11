# Minggu 10 — Modul 02: Incident CPU Spike

> **"CPU usage tiba-tiba 100%, latency naik, tapi jumlah request stabil."**

CPU spike adalah incident paling membingungkan karena **banyak sekali kemungkinan root cause** — dari yang jinak (traffic naik memang) sampai yang serius (infinite loop). Tidak bisa di-handle dengan `kubectl describe pod` saja, butuh korelasi dengan **Mimir, Tempo, dan profiling**.

---

## 🎯 Tujuan Modul

1. Membedakan **3 jenis CPU spike**: legitimate traffic, GC pressure, infinite loop
2. Membaca **CPU throttling metrics** di Mimir
3. Menggunakan **pprof** untuk Go app atau **perf** untuk profiling sistem
4. Menggunakan **Tempo** untuk identifikasi span yang lambat karena CPU
5. Menerapkan **HPA + CPU limits tuning** untuk prevention

---

## 🧠 Tiga Penyebab CPU Spike

```mermaid
mindmap
  root((CPU Spike))
    Legitimate Traffic
      User banyak
      Batch job jalan
      Scheduled task
    Runtime Overhead
      GC pressure
        Go GC terlalu sering
        JVM long GC pause
        Memory leak -> GC spike
      Connection pool exhausted
        Thread stuck waiting
    Code Bug
      Infinite loop
      O(n²) algorithm
      Regex catastrophic backtracking
      JSON marshal recursion
```

### Perbandingan Cepat

| Penyebab | CPU naik? | Latency naik? | Request naik? | Solusi |
|---|---|---|---|---|
| **Legitimate traffic** | ✅ | ✅ | ✅ | Scale up |
| **GC pressure** | ✅ | ✅ | ❌ | Tune GC, fix leak |
| **Infinite loop** | ✅ | ✅ (spike) | ❌ (mungkin turun karena stuck) | Patch code |

**Kunci diagnosis:** Lihat apakah **request naik atau tidak**. Kalau request naik → kemungkinan traffic. Kalau request tetap → cek GC atau bug code.

---

## 📦 Simulasi — CPU Spike dengan Stress Tool

### File: `manifests/01-cpu-spike-stress.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpu-spike-app
  namespace: insiden-lab
  labels:
    app: cpu-spike-app
    lab: insiden-m10-cpu
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cpu-spike-app
  template:
    metadata:
      labels:
        app: cpu-spike-app
    spec:
      containers:
      - name: burn
        image: polinux/stress:1.2.3
        args:
        - --cpu
        - "4"
        - --timeout
        - "1800s"
        resources:
          requests:
            cpu: 500m
            memory: 128Mi
          limits:
            cpu: "1"     # Limit cuma 1 core tapi app minta 4 → throttling!
            memory: 256Mi
```

### Deploy & Observe

```bash
kubectl apply -f manifests/01-cpu-spike-stress.yaml

kubectl top pods -n insiden-lab -l app=cpu-spike-app
# NAME                           CPU(cores)   MEMORY(bytes)
# cpu-spike-app-xxx              998m         50Mi     ← CPU mentok di limit (998m ≈ 1 core)
```

---

## 🔍 Symptoms

### Symptom 1 — Alert `ContainerCPUThrottling` Firing

(Alert yang kamu setup di Minggu 8):

```
[FIRING] ContainerCPUThrottling
  namespace  = insiden-lab
  pod        = cpu-spike-app-xxx
  container  = burn
  cpu_usage  = 998m (mendekati limit 1000m)
  throttled  = 60% periode terakhir throttled
  severity   = warning
```

### Symptom 2 — Dashboard CPU Panel Naik Tajam

Buka dashboard **"Cluster Resource Usage"** dari Minggu 8:
- Panel "CPU Usage by Pod" → pod `cpu-spike-app` grafik mendatar di 1000m (limit)
- Panel "CPU Throttling" → counter naik

### Symptom 3 — Latency Spike (kalau app kamu yang spike)

Buka dashboard **"Go App RED"**:
- Panel latency p95 → spike tiba-tiba
- Request rate → **stabil** (kunci diagnosis: bukan traffic)

---

## 🕵️ Investigation — Multi-Source Correlation

### Step 1 — Mimir: Cek CPU Usage vs Limit

```promql
# CPU usage (cores)
rate(container_cpu_usage_seconds_total{
  namespace="insiden-lab",
  pod="cpu-spike-app-xxx"
}[5m])

# CPU limit (cores)
kube_pod_container_resource_limits{
  namespace="insiden-lab",
  resource="cpu"
}

# Throttling percentage
rate(container_cpu_cfs_throttled_seconds_total[5m]) /
rate(container_cpu_cfs_throttled_periods_total[5m])
```

**Interpretasi hasil:**

| Pola | Arti |
|---|---|
| usage mendekati limit, throttling tinggi | CPU-bound, limit terlalu rendah |
| usage tinggi tapi throttling rendah | Pod punya CPU lebih dari limit (salah konfigurasi?) |
| usage tinggi tapi request count tidak naik | Bug code atau GC, bukan traffic |

### Step 2 — Mimir: Cek Request Rate

```promql
# Request per detik
sum(rate(http_requests_total{
  namespace="insiden-lab",
  service="cpu-spike-app"
}[5m]))
```

**Interpretasi:**
- Request rate **naik** → CPU spike karena traffic (solusi: scale up)
- Request rate **stabil** → CPU spike karena app issue (lanjut step 3)

### Step 3 — Tempo: Cek Span Distribution

Buka Tempo Explore:

```traceql
{ resource.service.name = "cpu-spike-app" } | select(span.duration)
```

**Interpretasi span:**
- Semua span **normal duration** (~50ms) → app sehat, CPU spike dari proses lain
- Span **lama semua** (~2-5 detik) → app sendiri yang lambat (lanjut step 4)
- Span ada yang **cepat, ada yang lambat** → ada endpoint tertentu yang bermasalah

### Step 4 — kubectl exec + top di dalam pod

```bash
# Masuk ke dalam pod
kubectl exec -it -n insiden-lab <pod-name> -- /bin/sh

# Lihat proses yang paling banyak pakai CPU
top -bn1 | head -20
# Output (contoh):
#   PID  USER   PR  NI  VIRT   RES   SHR  S  %CPU  %MEM   TIME+   COMMAND
#     1  root   20   0  2368   768   704  R  99.7  0.0   5:23.45  stress
#    42  root   20   0  2368   768   704  R  99.5  0.0   5:21.32  stress
#    43  root   20   0  2368   768   704  R  99.5  0.0   5:21.30  stress
#    44  root   0  -20  2368   768   704  R  99.5  0.0   5:21.29  stress
```

Semua proses pakai 99% CPU → **expected untuk stress test**. Tapi di production, kalau ada 1 proses yang 99%, dialah suspectnya.

### Step 5 — Profil CPU (Go app pakai pprof)

Kalau Go app:

```bash
# 1. Enable pprof di app (biasanya ada di endpoint /debug/pprof)
# Pastikan main.go punya:
# import _ "net/http/pprof"
# go http.ListenAndServe("localhost:6060", nil)

# 2. Port-forward ke pod
kubectl port-forward -n insiden-lab <pod-name> 6060:6060 &

# 3. Capture 30 detik CPU profile
curl -o cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30

# 4. Analisis
go tool pprof -top -cum cpu.prof
# Output menunjukkan fungsi mana yang paling banyak makan CPU
```

**Contoh output:**
```
Showing nodes accounting for 4500ms, 90% of 5000ms total
      flat  flat%   sum%        cum   cum%
    1500ms 30.00% 30.00%     4500ms 90.00%  github.com/user/app/json.Marshal
         0     0% 30.00%     3000ms 60.00%  github.com/user/app/handler.GetOrders
         0     0% 30.00%     1500ms 30.00%  runtime.gcDrain
```

**Insight:** 90% CPU di `json.Marshal` → ada payload yang sangat besar!

### Step 6 — Flame Graph (visualisasi)

```bash
# Install go tool pprof dengan web UI
go tool pprof -http=:8080 cpu.prof
# Buka http://localhost:8080 di browser → flame graph
```

Flame graph membuat visualisasi hierarchical: bar yang lebar = fungsi yang makan CPU banyak. Yang paling lebar di paling bawah = root cause.

---

## 🎯 Root Cause Analysis

### Skenario 1 — Legitimate Traffic Spike

**Bukti:**
- Request rate naik signifikan di Mimir
- Tempo span distribution normal (tidak ada outlier)
- Latency naik proporsional dengan request count

**Contoh:** Launch promo → user naik 10× → CPU naik.

### Skenario 2 — GC Pressure (Go / JVM)

**Bukti:**
- Request rate stabil
- GC time naik di Mimir (`go_gc_duration_seconds`)
- Heap allocation naik terus
- pprof menunjukkan `runtime.gcDrain` atau `runtime.scanobject` di top

**Contoh:** Code alokasi banyak short-lived object (string concatenation di loop) → GC harus kerja keras.

### Skenario 3 — Infinite Loop atau O(n²) Algorithm

**Bukti:**
- Request rate stabil atau bahkan turun
- 1 endpoint tertentu yang lambat (Tempo span distribution skewed)
- pprof menunjukkan 1 fungsi dominan (90%+ CPU)

**Contoh:** Loop `for i := 0; i < n; i++ { for j := 0; j < n; j++ { ... } }` dengan n = 10000 → 100 juta iterasi.

---

## 🛠️ Mitigation

### Opsi A — Scale Up (untuk skenario 1)

```bash
# Horizontal: tambah replicas
kubectl scale deployment/cpu-spike-app -n insiden-lab --replicas=4

# Vertical: naikkan CPU limit (kalau pakai HPA, jangan set limit)
kubectl set resources deployment/cpu-spike-app -n insiden-lab   --limits=cpu=2000m
```

### Opsi B — Tune GC (untuk skenario 2 — Go)

```bash
# Environment variable untuk tune GC
env:
- name: GOGC
  value: "200"          # default 100. Naikkan = GC lebih jarang, hemat CPU tapi pakai memory lebih
- name: GOMEMLIMIT
  value: "500MiB"       # Go 1.19+, hard memory limit
- name: GODEBUG
  value: "gctrace=1"    # Print GC stats ke stderr (untuk monitoring)
```

Atau optimize code:
```go
// ❌ BAD: alokasi string baru setiap iterasi
result := ""
for _, s := range items {
  result += s + ","  // alokasi string baru setiap kali
}

// ✅ GOOD: pakai strings.Builder
var sb strings.Builder
for _, s := range items {
  sb.WriteString(s)
  sb.WriteString(",")
}
result := sb.String()
```

### Opsi C — Kill Pod (untuk skenario 3, hot fix)

```bash
# Rollout restart untuk fresh pod
kubectl rollout restart deployment/cpu-spike-app -n insiden-lab

# Atau delete pod specific
kubectl delete pod -n insiden-lab -l app=cpu-spike-app
```

### Opsi D — Disable Endpoint (kalau bisa isolasi)

```bash
# Tambahkan annotation untuk feature flag (kalau app support)
kubectl annotate deployment/cpu-spike-app -n insiden-lab   feature-flag-endpoint-getorders=false
```

---

## 🛡️ Prevention

### Prevention 1 — HPA dengan CPU Metric

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: go-app-hpa
  namespace: produksi
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: go-app
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70   # Scale jika CPU > 70%
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 30
    scaleDown:
      stabilizationWindowSeconds: 300
```

**Cara kerja:** K8s otomatis tambah replicas saat CPU > 70% selama 30 detik.

### Prevention 2 — CPU Limit yang Realistis

**Prinsip:** Untuk app yang auto-scale (HPA), **JANGAN set CPU limit** (atau set sangat tinggi). Limit menyebabkan throttling yang malah memperlambat app.

```yaml
spec:
  containers:
  - name: go-app
    resources:
      requests:
        cpu: 500m       # Scheduler tahu butuh minimal ini
      # NO limits! Biarkan pod pakai CPU yang tersedia
```

Atau kalau limit WAJIB (misal: cluster shared), set 2× dari peak usage:
```yaml
limits:
  cpu: "2"            # peak observed 1 core, kasih 2× buffer
```

### Prevention 3 — pprof Endpoint Selalu Aktif (untuk Go app)

```go
// main.go
import (
  "net/http"
  _ "net/http/pprof"  // Register pprof handlers
)

func main() {
  // pprof di port terpisah (jangan di-expose ke public!)
  go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
  }()
  
  // Main app
  // ...
}
```

**PENTING:** Port pprof TIDAK boleh di-expose ke internet! Hanya lewat `kubectl port-forward`.

### Prevention 4 — CPU Profiling di CI (load test)

```yaml
# .gitlab-ci.yml
load-test:
  stage: test
  script:
  - |
    # Jalankan app
    docker run -d --name test-app -p 6060:6060 -p 8080:8080 my-app:latest
    
    # Capture CPU profile selama load test
    hey -z 60s -c 50 http://localhost:8080/api/orders &
    sleep 60
    curl -o /tmp/cpu.prof http://localhost:6060/debug/pprof/profile?seconds=30
    
    # Validasi CPU tidak hot
    go tool pprof -top /tmp/cpu.prof > /tmp/cpu.txt
    if grep -q "90.00%" /tmp/cpu.txt; then
      echo "ERROR: One function uses 90%+ CPU"
      exit 1
    fi
  allow_failure: false
```

### Prevention 5 — Alert untuk CPU Throttling

(Alert yang sudah kamu buat di Minggu 8):

```yaml
- alert: ContainerCPUThrottling
  expr: |
    rate(container_cpu_cfs_throttled_seconds_total[5m]) > 0.5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Container {{ $labels.container }} throttled > 50% of CPU time"
```

---

## 🧪 Verifikasi Recovery

```bash
# 1. CPU usage kembali normal
kubectl top pods -n insiden-lab -l app=cpu-spike-app
# NAME                           CPU(cores)   MEMORY(bytes)
# cpu-spike-app-xxx              50m          50Mi      ← turun

# 2. Throttling hilang
# Cek di Mimir: rate(container_cpu_cfs_throttled_seconds_total) == 0

# 3. Latency recovered
# Cek dashboard Go App RED: p95 kembali ke baseline

# 4. Alert resolved
# Cek Slack: [RESOLVED] ContainerCPUThrottling
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/01-cpu-spike-stress.yaml
```

---

## 📖 Rangkuman

| Aspek | Catatan |
|---|---|
| **Symptom utama** | CPU usage = limit, latency naik, throttling tinggi |
| **Cara diagnosis** | Cek request rate → traffic atau bukan? Cek GC → GC pressure? pprof → infinite loop? |
| **3 root cause** | Traffic naik, GC pressure, infinite loop |
| **Mitigation** | Scale, tune GC, kill pod |
| **Prevention terbaik** | HPA + realistic limits + pprof di CI |
| **Alert firing** | `ContainerCPUThrottling` (Minggu 8) ✅ |

---

## ➡️ Modul 03: Memory Leak

CPU spike naik **turun dengan cepat** (bisa solve dalam menit). Sekarang kita masuk **memory leak** — naik **perlahan tapi pasti**, dan kalau tidak di-handle, baru explode jadi OOMKilled setelah berjam-jam/hari.

👉 Lanjut ke `03-incident-memory-leak.md`
