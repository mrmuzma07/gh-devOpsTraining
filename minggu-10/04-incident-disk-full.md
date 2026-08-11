# Incident #4 — Disk Pressure & Full Filesystem

> **Satu kalimat:** Disk node penuh → kubelet men‑evict Pod → Deployment Pending → aplikasi down bukan karena aplikasinya rusak, tapi karena **node tempat dia ingin berjalan tidak punya ruang**.

Bayangkan laptop Anda: hard disk penuh, Anda tidak bisa install aplikasi baru atau bahkan update file apa pun. Kubernetes punya masalah yang sama — kalau disk di node penuh, dia bilang "tidak ada tempat untuk Pod baru, evict yang sudah ada".

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:

1. Membedakan **ephemeral storage** (container writable layer) vs **node filesystem** (host disk).
2. Mengenali tanda‑tanda **DiskPressure** di event Kubernetes.
3. Menggunakan `df`, `du`, dan `ncdu` untuk mencari "siapa yang makan disk".
4. Membuat strategi **pencegahan** lewat log rotation, image GC, dan PVC limit.
5. Memonitor `node_filesystem_avail_bytes` lewat Mimir + Alert.

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari sisi user
- Aplikasi tiba‑tiba return **503 Service Unavailable** atau timeout.
- Dashboard Prometheus: panel "Available Replicas" tiba‑tiba turun drastis.
- Alert: **KubeNodeNotReady** atau **KubeDeploymentReplicasMismatch**.

### 1.2 Dari sisi kubectl
```bash
$ kubectl get nodes
NAME           STATUS                     ROLES                  AGE   VERSION
laptop-node-1  Ready,SchedulingDisabled   control-plane,master   30d   v1.29.0
```

Tanda kritis: `Ready,SchedulingDisabled` — node masih hidup tapi tidak mau menerima Pod baru.

```bash
$ kubectl describe node laptop-node-1 | grep -A 5 Conditions
Conditions:
  Type                 Status  Reason          Message
  ----                 ------  ------          -------
  Ready                False   DiskPressure    kubelet has disk pressure
  MemoryPressure       False   <nil>
  PIDPressure          False   <nil>
```

**`DiskPressure = True`** adalah kunci diagnosis.

### 1.3 Dari Mimir (PromQL)
```promql
# % disk terpakai
100 - (node_filesystem_avail_bytes{mountpoint="/"} 
       / node_filesystem_size_bytes{mountpoint="/"} * 100)
# Hasil: 95.2% — sudah melewati threshold 85% (image GC) & 95% (eviction)
```

### 1.4 Dari Loki
```logql
{kubernetes_pod=~".+"} |= "evicted" | json
# Pesan: "Pod ephemeral local storage usage exceeds the limit"
```

---

## 2. 🧠 7-Layer Drill-Down untuk Disk Penuh

```
┌─────────────────────────────────────────────────────┐
│ Layer 7 — UX                                         │  ← 503 dari user
├─────────────────────────────────────────────────────┤
│ Layer 6 — App (Container Process)                    │  ← ENOSPC: no space left on device
├─────────────────────────────────────────────────────┤
│ Layer 5 — Container Runtime (containerd writable)    │  ← EmptyDir overflow, log files
├─────────────────────────────────────────────────────┤
│ Layer 4 — OS Layer (host /var/lib/docker)            │  ← Image layers, stopped containers
├─────────────────────────────────────────────────────┤
│ Layer 3 — Kernel                                     │  ← Inode habis (terlalu banyak file kecil)
├─────────────────────────────────────────────────────┤
│ Layer 2 — Node Filesystem                            │  ← df -h: Use% 98%
├─────────────────────────────────────────────────────┤
│ Layer 1 — Hardware                                   │  ← Disk 100GB, kepakai 99GB
└─────────────────────────────────────────────────────┘
```

**Untuk insiden ini, fokus di Layer 4 (OS) dan Layer 5 (Container Runtime).**

---

## 3. 🔍 Investigasi Step-by-Step

### Step 1 — Identifikasi node mana yang penuh
```bash
$ kubectl top node
NAME           CPU(cores)  CPU%  MEMORY(bytes)  MEMORY%
laptop-node-1  234m        11%   2150Mi         54%

# top node tidak menampilkan disk! Harus pakai custom metric
$ kubectl get --raw /api/v1/nodes/laptop-node-1/proxy/stats/summary | jq '.node.fs'
{
  "availableBytes": 2147483648,   # 2GB tersisa
  "capacityBytes": 107374182400,  # 100GB total
  "usedBytes": 105226698752       # 98GB terpakai
}
```

Atau lebih simpel lewat **node-exporter** yang sudah kita install di Week 5:
```bash
$ curl -s http://laptop-node-1:9100/metrics | grep node_filesystem_avail
node_filesystem_avail_bytes{mountpoint="/"} 2.147e+09
```

### Step 2 — Cari tahu direktori mana yang makan disk
```bash
# SSH ke node (di k3s single-node = laptop kita sendiri)
$ df -h /
Filesystem      Size  Used Avail Use% Mounted on
/dev/disk1s5   100Gi  98Gi  2.0Gi  98% /

# Cari direktori terbesar di level atas
$ sudo du -sh /* 2>/dev/null | sort -hr | head -10
48G   /var
22G   /Users
12G   /Applications
6.2G  /private

# Terus drill-down ke /var
$ sudo du -sh /var/* 2>/dev/null | sort -hr | head -10
30G   /var/lib             ← containerd/ docker image layers
8.5G  /var/log             ← host logs (kubelet, containerd)
1.2G  /var/lib/kubelet     ← Pod sandbox + volumes
```

### Step 3 — Cek image & stopped container
```bash
# Lihat image apa saja yang ada
$ sudo crictl images
IMAGE                                    TAG     SIZE
docker.io/library/redis                  latest  104MB
docker.io/library/postgres               15      376MB
docker.io/library/nginx                  1.27    142MB
docker.io/yourname/buggy-app             v1      680MB    ← culprit
docker.io/yourname/buggy-app             v2      712MB    ← culprit
docker.io/yourname/buggy-app             v3      728MB
docker.io/yourname/buggy-app-debug       latest  1.2GB    ← debug image besat
...

# Total image cache
$ sudo du -sh /var/lib/containerd
28G   /var/lib/containerd
```

### Step 4 — Cek log yang menumpuk
```bash
# Cari log file > 100MB
$ sudo find /var/log -type f -size +100M -exec ls -lh {} \;
-rw-r--r-- 1 root root 4.2G  /var/log/containers/buggy-app-xxx.log   ← culprit!
-rw-r--r-- 1 root root 1.8G  /var/log/pods/xxx/0.log
```

### Step 5 — Cek inode (sering lupa)
```bash
$ df -i /
Filesystem      Inodes  IUsed   IFree IUse% Mounted on
/dev/disk1s5    6.2M    6.1M    100K  98%   /    ← INODE HABIS!
```

> **Insight:** `df -h` bisa masih menunjukkan sisa ruang (mis. 2GB), tapi inode habis. Ini terjadi saat ada jutaan file kecil. Ciri umum: Pod `PodInitializing` lama, image pull gagal dengan error "no space".

---

## 4. 🎯 Root Cause Analysis

### Skenario A — Image Cache Menumpuk
**Penyebab:** Deployment sering di-redeploy tanpa membersihkan image lama. Setiap `docker build --tag app:vN` membuat layer baru yang disimpan di `/var/lib/containerd`.

**Gejala spesifik:**
- `du -sh /var/lib/containerd` > 20GB
- Image lama masih ada walau Pod sudah dihapus
- Container image GC tidak pernah jalan

### Skenario B — Log Container Tidak Dirotasi
**Penyebab:** Aplikasi logging sangat verbose (level=DEBUG) dan stdout/stderr di-redirect ke file di `/var/log/containers/`. Logrotate containerd default = 10MB per file, tapi bisa gagal kalau Pod sering restart.

**Gejala spesifik:**
- File `/var/log/containers/*.log` > 1GB
- 1 Pod single line JSON per detik = 86.400 baris/hari = ~50MB

### Skenario C — EmptyDir / ephemeral-storage Overflow
**Penyebab:** Aplikasi menulis file temporer ke `/tmp` atau `/cache` di dalam container. EmptyDir default **tidak ada limit**.

**Gejala spesifik:**
- Event Pod: `Pod ephemeral local storage usage exceeds the limit`
- App crash dengan `ENOSPC: no space left on device`

### Skenario D — PVC Penuh (lihat juga insiden #6)
**Penyebab:** Database/data log file mengisi PersistentVolume sampai 100%. `kubelet` tidak bisa mount PVC baru.

### Skenario E — Inode Habis
**Penyebab:** Aplikasi membuat banyak file kecil (mis. session file per request) tanpa pernah dihapus.

---

## 5. 🛠️ Mitigasi (Bertahap)

### Tahap 1 — Buang sampah yang aman dibuang (≤ 5 menit)

```bash
# 1. Hapus image dangling (tanpa tag)
$ sudo crictl rmi --prune
# Membebaskan ~5-15GB biasanya

# 2. Hapus stopped container (kalau ada)
$ sudo ctr -n k8s.io containers ls | grep "Exited"
$ sudo ctr -n k8s.io containers rm <container-id>

# 3. Compress logfile aktif yang kebesaran (TANPA menghapus!)
$ sudo truncate -s 0 /var/log/containers/buggy-app-xxx.log
# truncate -s 0 = kosongkan file, tapi inode tetap, jadi tidak ganggu proses yang sedang menulis

# 4. Docker/containerd system prune
$ sudo crictl images -q | xargs -L1 sudo crictl rmi 2>/dev/null  # HATI-HATI!
```

### Tahap 2 — Kembalikan node ke Ready (≤ 15 menit)

```bash
# 1. Disable scheduling dulu supaya Pod baru tidak langsung masuk
$ kubectl cordon laptop-node-1

# 2. Force image GC manual
$ sudo /usr/bin/containerd-prune

# 3. Atau restart containerd (ekstrem, evict semua Pod)
$ sudo systemctl restart containerd
# Ini evict semua Pod di node, tapi setelah containerd up, scheduler akan reschedule.

# 4. Uncordon setelah disk turun
$ kubectl uncordon laptop-node-1
$ watch kubectl get nodes
# Tunggu STATUS kembali ke Ready
```

### Tahap 3 — Tangani aplikasi yang sedang ditulis (ephemeral-storage)

```bash
# Login ke Pod yang masih hidup
$ kubectl exec -it buggy-app-xxx -- /bin/sh

# Cek penggunaan EmptyDir
$ df -h /tmp
Filesystem      Size  Used Avail Use% Mounted on
/dev/...        1.0G  980M  20M  98% /tmp

# Hapus file temporer yang aman
$ rm -rf /tmp/cache/*
```

---

## 6. 🛡️ Pencegahan Jangka Panjang

### 6.1 Set limit ephemeral storage di Pod
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: buggy-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: yourname/buggy-app:v3
        resources:
          requests:
            ephemeral-storage: "1Gi"
          limits:
            ephemeral-storage: "2Gi"   # Pod akan di-evict jika lewat 2GB
        volumeMounts:
        - name: cache
          mountPath: /tmp
      volumes:
      - name: cache
        emptyDir:
          sizeLimit: 500Mi             # Hard limit untuk EmptyDir
```

### 6.2 Konfigurasi kubelet image GC (node-level)
```bash
# /etc/rancher/k3s/config.yaml (k3s) atau /var/lib/kubelet/config.yaml
kubelet-arg:
  - "image-gc-high-threshold=85"   # mulai GC saat 85%
  - "image-gc-low-threshold=80"    # stop GC saat 80%
  - "minimum-image-ttl-duration=10m"  # image baru tidak boleh langsung dihapus dalam 10 menit
```

Setelah restart kubelet:
```bash
$ sudo systemctl restart k3s
```

### 6.3 Logrotate untuk container log
Buat ConfigMap:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: container-log-rotate
  namespace: kube-system
data:
  logrotate.conf: |
    /var/log/containers/*.log {
      size 50M
      rotate 5
      compress
      missingok
      notifempty
      copytruncate
    }
```
Pakai CronJob atau DaemonSet yang jalan tiap jam.

### 6.4 Alerts di Mimir (PrometheusRule)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: disk-alerts
  namespace: monitoring
spec:
  groups:
  - name: disk
    rules:
    - alert: NodeFilesystemAlmostOutOfSpace
      expr: |
        (node_filesystem_avail_bytes{mountpoint="/"} 
         / node_filesystem_size_bytes{mountpoint="/"} * 100) < 15
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Node {{ $labels.instance }} disk usage {{ $value | printf \"%.1f\" }}%"

    - alert: NodeFilesystemWillFillIn4Hours
      expr: |
        predict_linear(node_filesystem_avail_bytes{mountpoint="/"}[2h], 4*3600) < 0
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Node {{ $labels.instance }} akan penuh dalam 4 jam"

    - alert: NodeInodesExhausted
      expr: node_filesystem_files_free{mountpoint="/"} < 100000
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Inode hampir habis di {{ $labels.instance }}"

    - alert: KubeletDiskPressure
      expr: kube_node_status_condition{condition="DiskPressure",status="true"} == 1
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "Node {{ $labels.node }} dalam kondisi DiskPressure"
```

### 6.5 Capacity Planning
```
Estimasi ukuran deployment:
  container image   : 700 MB × 3 replicas × 2 versi (lama + baru) = 4.2 GB
  container logs    : 50 MB/hari × 365 hari × 5 Pod = 91 GB ← BAHAYA!
  ephemeral storage : 500 MB × 10 Pod = 5 GB
  PVC               : 50 GB (Postgres data)
  Total estimasi    : ~150 GB

Rekomendasi: minimal 250 GB SSD untuk laptop SRE
```

---

## 7. 🧪 Lab Hands-On: Reproduksi Disk Penuh

### Skenario Lab
Kita akan memenuhi disk node dengan file besar, lalu lihat efeknya ke Pod baru.

### Step 1 — Buat manifest stress disk
File: `minggu-10/manifests/04-disk-fill.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: disk-fill
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: disk-fill
  template:
    metadata:
      labels:
        app: disk-fill
    spec:
      containers:
      - name: filler
        image: alpine:3.19
        command: ["/bin/sh", "-c"]
        args:
        - |
          echo "Mulai mengisi disk..."
          # Buat 1 file 1GB, diulang sampai disk penuh
          while true; do
            dd if=/dev/zero of=/data/bigfile-$(date +%s).bin bs=1M count=1024 2>/dev/null
            sleep 5
          done
        resources:
          requests:
            ephemeral-storage: "1Gi"
            cpu: 50m
            memory: 64Mi
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        emptyDir:
          sizeLimit: "50Gi"  # Batas 50GB, lebih dari ini akan evict
```

### Step 2 — Deploy dan amati
```bash
$ kubectl apply -f manifests/04-disk-fill.yaml
deployment.apps/disk-fill created

$ kubectl get pods -w
NAME                          READY   STATUS    RESTARTS   AGE
disk-fill-7c8b9d5f8-xxx       1/1     Running   0          5s

# Tunggu ~1 menit, Pod akan di-evict
$ kubectl get pods -w
disk-fill-7c8b9d5f8-xxx       0/1   Evicted   0          67s
disk-fill-7c8b9d5f8-xxx       0/1   Pending   0          0s   # re-created tapi Pending

$ kubectl describe pod disk-fill-7c8b9d5f8-xxx | grep -A 5 Events
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  5s    default-scheduler  0/1 nodes are available: 1 Insufficient ephemeral-storage.
```

### Step 3 — Investigasi node
```bash
$ kubectl describe node laptop-node-1 | grep -A 8 Conditions
Conditions:
  Type                 Status  Reason          Message
  ----                 ------  ------          -------
  Ready                False   DiskPressure    kubelet has disk pressure

$ kubectl top node  # CPU/Mem masih normal
NAME           CPU(cores)  CPU%  MEMORY(bytes)  MEMORY%
laptop-node-1  50m         2%    800Mi          20%

# Tapi disk sudah penuh
$ df -h /
Filesystem      Size  Used Avail Use% Mounted on
/dev/disk1s5   100Gi  100Gi  50M  100% /

# Cek apa yang makan disk
$ sudo du -sh /var/lib/containerd  # 8GB image cache normal
$ sudo du -sh /var/lib/kubelet/pods/*/volumes  # banyak EmptyDir dari Pod yang di-evict
```

### Step 4 — Hapus stressor
```bash
$ kubectl delete -f manifests/04-disk-fill.yaml
deployment.apps "disk-fill" deleted

# Bersihkan EmptyDir yang masih ada
$ sudo rm -rf /var/lib/kubelet/pods/*/volumes/kubernetes.io~empty-dir/data/*

# Uncordon
$ kubectl uncordone laptop-node-1
```

### Step 5 — Verifikasi dengan Loki & Tempo
```logql
# Cari event eviksi
{kubernetes_pod=~"disk-fill.*"} |= "evicted" | json | line_format "{{.reason}}: {{.message}}"
```

```text
Evicted: Pod ephemeral local storage usage exceeds the limit
```

Tempo trace: cari trace Pod yang di-evict, akan ada span `kubelet.evictPod` dengan atribut `reason=ephemeralStorage`.

---

## 8. 📋 Cheat Sheet Diagnosis

```text
GEJALA                 | CEK PERTAMA              | PENYEBAB UMUM
-----------------------|--------------------------|---------------------
Node NotReady          | kubectl describe node    | DiskPressure=True
Pod Pending forever    | describe pod → Events    | Insufficient ephemeral-storage
App error ENOSPC       | df -h di dalam container | EmptyDir full
df -i 100%             | df -i /                  | Terlalu banyak file kecil
Image pull lambat      | crictl images | size     | Image cache besar
Log sangat besar       | ls -lh /var/log/...     | Log rotation off
PVC Pending            | describe pvc             | Penyebab lain (lihat insiden #6)
```

---

## 9. 🔗 Korelasi Multi-Sumber

```
Mimir (kapan): node_filesystem_avail_bytes turun gradual sejak 3 hari lalu
   ↓
Loki (apa): container log buggy-app error "no space left on device"
   ↓
Tempo (di mana): kubelet.evictPod span di node-1, durasi 47ms
   ↓
kubectl (state): node condition DiskPressure=True
   ↓
ROOT CAUSE: image GC tidak jalan, ada 8 image lama di cache = 12GB
```

---

## 10. ✅ Rangkuman

| Aspek | Pelajaran |
|---|---|
| **Gejala utama** | Node NotReady, Pod Pending, image pull lambat |
| **Diagnosis cepat** | `df -h`, `du -sh /*`, `crictl images` |
| **Mitigasi darurat** | `crictl rmi --prune`, truncate log, restart containerd |
| **Pencegahan** | ephemeral-storage limit, logrotate, image GC tuning, capacity planning |
| **Monitoring** | `node_filesystem_avail_bytes` + `predict_linear` |

---

## 11. ✏️ Latihan Mandiri

1. Buat **PromQL query** untuk mendeteksi folder yang pertumbuhannya > 1GB/hari.
2. Tulis **helm values** untuk `kube-prometheus-stack` yang tune image GC threshold.
3. Simulasikan **inode habis**: deploy Pod yang membuat 1 juta file kecil, lalu observasi.
4. Buat **runbook** di Confluence/Notion untuk tim on-call.

---

**Lanjut ke insiden berikutnya:** [05-incident-dns-error.md](./05-incident-dns-error.md) — saat Pod tidak bisa saling bicara karena DNS rusak.
