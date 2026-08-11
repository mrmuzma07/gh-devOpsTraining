# Modul 03 — Capacity Planning: Merancang Infrastruktur yang Cukup (Tidak Lebih, Tidak Kurang)

> **Satu kalimat:** Capacity Planning adalah seni **memprediksi kebutuhan resource di masa depan** berdasarkan data historis, sehingga Anda tidak over-provisioning (buang uang) atau under-provisioning (downtime).

Bayangkan Anda buka restoran. Anda harus memutuskan: berapa kursi? berapa bahan baku? berapa karyawan? Kalau terlalu sedikit, pelanggan kecewa & pergi. Kalau terlalu banyak, Anda rugi. **Capacity Planning** untuk infrastruktur adalah persis seperti itu — keputusan investasi berdasarkan prediksi permintaan.

---

## 🎯 Learning Outcomes

1. Memahami **Capacity Planning** sebagai disiplin ilmu SRE (bukan sekadar "tambah server kalau lambat").
2. Menghitung **headroom**, **growth rate**, dan **forecast horizon**.
3. Menggunakan **Mimir `predict_linear()`** untuk memprediksi kapan resource habis.
4. Melakukan **right-sizing** Pod resources berdasarkan data historis (VPA).
5. Mendesain **capacity model** untuk Mini Production Platform Anda.

---

## 1. 🧠 Tiga Pertanyaan Capacity Planning

Setiap keputusan capacity planning HARUS menjawab tiga pertanyaan:

```text
┌──────────────────────────────────────────────────────────────────┐
│  Tiga Pertanyaan Capacity Planning                              │
│                                                                  │
│  1. BERAPA resource yang saya butuhkan?                           │
│     → CPU cores, RAM GB, Disk GB, Network Mbps                   │
│                                                                  │
│  2. KAPAN saya membutuhkannya?                                    │
│     → Bulan depan? Quarter depan? Tahun depan?                   │
│                                                                  │
│  3. APA yang akan terjadi kalau saya salah?                       │
│     → Over-provision: rugi $X/bulan                              │
│     → Under-provision: downtime, customer churn                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. 📊 Capacity Model: Formula Sederhana

### 2.1 Availability Capacity (CPU & Memory)

```text
Required Capacity = Current Peak Load × Growth Factor × Headroom
```

**Variabel:**
- **Current Peak Load** = 95th percentile (p95) konsumsi CPU/Memory dari metrics.
- **Growth Factor** = (1 + projected_growth_rate)^months
- **Headroom** = buffer 25-50% untuk insiden atau lonjakan tak terduga.

**Contoh Perhitungan:**

```text
Current p95 CPU usage:  800m (0.8 core) di bulan ini
Growth rate per bulan: 15% (karena campaign marketing Q1)
Headroom:              30% (buffer untuk insiden)
Forecast horizon:      3 bulan

Required CPU in 3 months = 0.8 × (1 + 0.15)^3 × 1.3
                        = 0.8 × 1.521 × 1.3
                        = 1.58 core

Rekomendasi: Provision 2 cores (headroom naik ke 27% dari 1.58).
```

### 2.2 Storage Capacity (Disk)

```text
Required Storage = Current Usage + (Daily Growth × Days) + Safety Reserve
```

**Contoh:**

```text
Current DB size:        50 GB
Daily growth:           500 MB/hari (1.5 GB/bulan)
Forecast horizon:       6 bulan
Safety reserve:         20%

Required storage in 6 months = 50 + (0.5 × 30 × 6) × 1.2
                            = 50 + 90 × 1.2
                            = 158 GB

Rekomendasi: Provision 200 GB storage class (longhorn / ceph).
```

### 2.3 Throughput Capacity (RPS)

```text
Required RPS = Peak RPS × Burst Factor
```

**Contoh:**

```text
Current peak RPS:    1,200 RPS (campaign 11.11 lalu)
Burst factor:        2.5x (kadang flash sale 3x lipat)
Target safety:       20% headroom

Required capacity = 1,200 × 2.5 / 1.2 = 2,500 RPS
```

---

## 3. 📈 Forecast dengan Mimir `predict_linear()`

PromQL punya fungsi `predict_linear()` yang melakukan **regresi linear** terhadap time-series dan memprediksi nilai masa depan.

### 3.1 Prediksi Kapan Disk Penuh

```promql
# Prediksi penggunaan disk 30 hari ke depan
predict_linear(
  node_filesystem_used_bytes{mountpoint="/"}[30d],
  30 * 24 * 3600   # 30 hari dalam detik
)
```

**Interpretasi:**
- Jika hasil `> node_filesystem_size_bytes` → **disk akan penuh dalam 30 hari**, alert.
- Jika hasil `> 80% size` → Anda punya 30 hari untuk tambah storage.

### 3.2 Alert Capacity Planning dengan predict_linear

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: capacity-planning-alerts
  namespace: monitoring
spec:
  groups:
  - name: capacity
    interval: 1h   # Tidak perlu re-evaluasi cepat, 1 jam cukup
    rules:
    - alert: DiskWillFillIn30Days
      expr: |
        predict_linear(
          node_filesystem_used_bytes{mountpoint="/"}[7d],
          30 * 24 * 3600
        ) > node_filesystem_size_bytes{mountpoint="/"}
      for: 6h
      labels:
        severity: warning
        team: infra
      annotations:
        summary: "Node {{ $labels.instance }} disk akan penuh dalam 30 hari"

    - alert: CPUSaturationForecast
      expr: |
        predict_linear(
          instance:node_cpu:ratio_rate5m[7d],
          14 * 24 * 3600
        ) > 0.85
      for: 6h
      labels:
        severity: warning
      annotations:
        summary: "CPU cluster akan > 85% dalam 14 hari"

    - alert: MemoryExhaustionForecast
      expr: |
        predict_linear(
          node_memory_MemAvailable_bytes[14d],
          7 * 24 * 3600
        ) < 0
      for: 12h
      labels:
        severity: critical
      annotations:
        summary: "Memory cluster akan HABIS dalam 7 hari!"
```

### 3.3 Visualisasi Forecast di Grafana

```json
{
  "title": "Capacity Forecast — Disk & Memory",
  "type": "timeseries",
  "targets": [
    {
      "expr": "predict_linear(node_filesystem_used_bytes{mountpoint=\"/\"}[7d], 30*24*3600)",
      "legendFormat": "Disk forecast 30d - {{instance}}"
    },
    {
      "expr": "node_filesystem_size_bytes{mountpoint=\"/\"}",
      "legendFormat": "Disk total - {{instance}}"
    }
  ],
  "fieldConfig": {
    "defaults": {
      "custom": {
        "fillOpacity": 10,
        "showPoints": "never"
      }
    }
  }
}
```

---

## 4. ⚖️ Right-Sizing dengan VPA (Vertical Pod Autoscaler)

### 4.1 Masalah: Kebanyakan Pod Over-Provisioned

Lihat data historis — kebanyakan tim set CPU requests = 1000m, padahal aktual usage cuma 200m. Ini **pemborosan 5x**.

```bash
# Cek rekomendasi VPA
$ kubectl describe vpa checkout-service-vpa
# Recommendation:
#   Container: app
#     Target:
#       Cpu:     150m      ← aktual kebutuhan
#       Memory:  180Mi
#     Upper Bound:
#       Cpu:     250m      ← maximum historically observed
#       Memory:  320Mi
#     Lower Bound:
#       Cpu:     80m       ← minimum historically observed
#       Memory:  100Mi
```

### 4.2 Deploy VPA

File: `minggu-11/manifests/03-vpa-checkout.yaml`

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: checkout-service-vpa
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: checkout-service
  updatePolicy:
    updateMode: "Auto"        # atau "Off" untuk rekomendasi saja
  resourcePolicy:
    containerPolicies:
    - containerName: app
      minAllowed:
        cpu: 50m
        memory: 64Mi
      maxAllowed:
        cpu: 1000m
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
```

### 4.3 Perbandingan: HPA vs VPA

| Aspek | HPA (Horizontal) | VPA (Vertical) |
|---|---|---|
| **Mekanisme** | Tambah/kurangi **jumlah** Pod | Adjust **CPU/Memory** per Pod |
| **Trigger** | CPU/Memory/custom metric | Data historis VPA recommender |
| **Waktu respons** | Detik (scale up) sampai menit (scale down) | Menit hingga jam (butuh restart Pod) |
| **Cocok untuk** | Stateless app, traffic spike | Stateful app, memory-intensive |
| **Jangan dipakai bareng** | - | Dengan HPA untuk metric CPU yang sama |

### 4.4 Anti-Pattern: Jangan Pakai HPA + VPA untuk Metric yang Sama

```text
❌ SALAH:
  HPA: target CPU 70%
  VPA: update CPU based on usage
  → VPA naikkan request → HPA scale down (karena usage < 70% dari request)
  → VPA turunkan request → HPA scale up
  → Infinite loop! 

✅ BENAR:
  HPA: target CPU 70% (untuk scale out saat traffic naik)
  VPA: target memory saja (untuk right-size memory)
  → Dua mekanisme tidak konflik
```

---

## 5. 💼 Capacity Planning untuk Mini Production Platform

### 5.1 Resource Inventory Saat Ini (Week 1-10)

```text
┌──────────────────────────────────────────────────────────────────┐
│  Mini Production Platform — Resource Inventory (Laptop k3s)     │
│                                                                  │
│  COMPONENT         REQUESTS    ACTUAL P95    HEADROOM            │
│  ─────────────────────────────────────────────────────────────   │
│  checkout-service  200m/256Mi  120m/180Mi    ✓ 40% OK            │
│  payment-service   300m/512Mi  280m/450Mi    ⚠  6% tight!       │
│  postgres          500m/1Gi    480m/950Mi    ❌ 4% critical     │
│  redis             100m/128Mi  80m/100Mi     ✓ 20% OK            │
│  prometheus        200m/512Mi  180m/480Mi    ✓ 10% OK            │
│  loki              100m/256Mi  90m/240Mi     ✓ 10% OK            │
│  tempo             100m/256Mi  60m/180Mi     ✓ 40% OK            │
│                                                                  │
│  REKOMENDASI AKSI:                                               │
│  1. payment-service: tambah replicas atau naikkan limit CPU.    │
│  2. postgres: naikkan request ke 1 core / 2Gi sebelum OOM.      │
└──────────────────────────────────────────────────────────────────┘
```

### 5.2 Capacity Model Spreadsheet (Template)

Buat spreadsheet `capacity-model.xlsx` di repo:

```text
Sheet 1: Resource Inventory
- Component name
- Replicas (current)
- CPU request / limit
- Memory request / limit
- Actual p95 (last 30 days)
- Headroom %
- Status (OK / Warning / Critical)

Sheet 2: Forecast (3-6-12 months)
- Component
- Current usage
- Growth rate (% monthly)
- Forecast 3M
- Forecast 6M
- Forecast 12M
- Action needed

Sheet 3: Cost Analysis
- Component
- Instance type
- Cost / month
- Cost per 1k RPS
- ROI calculation
```

### 5.3 Growth-Driven Capacity Add

**Kapan Anda harus scale up?** Berikut rule sederhana:

```text
Trigger untuk scale up CPU:
  Current p95 CPU > 70% selama 7 hari berturut-turut
  ATAU
  predict_linear(cpu[14d], 7d) > 80%

Trigger untuk scale up Memory:
  Current p95 memory > 75% selama 7 hari
  ATAU
  predict_linear(memory[14d], 7d) > 85%

Trigger untuk scale up Storage:
  predict_linear(disk_used[30d], 30d) > 80% capacity
  ATAU
  kubelet_volume_stats_used_bytes > 85%
```

---

## 6. 🧪 Lab Hands-On: Capacity Forecasting

### 6.1 Generate Traffic Pattern Realistis

Buat script untuk generate traffic naik gradual selama 1 jam:

File: `minggu-11/manifests/03-traffic-sim.yaml`

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: traffic-simulator
  namespace: default
spec:
  template:
    spec:
      containers:
      - name: curl
        image: alpine/curl
        command: ["/bin/sh", "-c"]
        args:
        - |
          # Generate traffic naik dari 10 RPS ke 100 RPS dalam 1 jam
          for minute in $(seq 0 59); do
            rps=$((10 + minute * 1.5))
            echo "Minute $minute: target $rps RPS"
            for i in $(seq 1 $rps); do
              curl -s -o /dev/null http://checkout-service:8080/ &
            done
            sleep 60
          done
          wait
      restartPolicy: Never
  backoffLimit: 1
```

### 6.2 Observe Trend di Mimir

```promql
# Prediksi CPU usage 1 jam ke depan
predict_linear(
  rate(container_cpu_usage_seconds_total{pod=~"checkout-service-.*"}[30m])[30m:1m],
  3600
)

# Prediksi memory usage 24 jam ke depan
predict_linear(
  container_memory_working_set_bytes{pod=~"checkout-service-.*"}[1h],
  24 * 3600
)
```

### 6.3 Buat Capacity Report Otomatis

Buat ServiceMonitor + Recording rule untuk quarterly report:

```yaml
- record: capacity:report:cpu_p95_30d
  expr: |
    quantile_over_time(0.95,
      rate(container_cpu_usage_seconds_total{namespace="default"}[5m])[30d:1h]
    )
  
- record: capacity:report:memory_p95_30d
  expr: |
    quantile_over_time(0.95,
      container_memory_working_set_bytes{namespace="default"}[5m])[30d:1h]
    )
```

---

## 7. 📋 Cheat Sheet: Capacity Planning Decisions

```text
ASPEK                | TOOL                    | OUTPUT
---------------------|-------------------------|---------------------------------------
Forecast CPU         | predict_linear()        | Kapan node jenuh
Forecast Memory      | predict_linear()        | Kapan OOM
Forecast Disk        | predict_linear()        | Kapan disk penuh
Right-size Pod       | VPA recommender         | CPU/Memory recommendation
Scale out            | HPA                     | Tambah replicas
Cost analysis        | Kubecost / manual calc  | $/month per service
Growth tracking      | Grafana Mimir trending  | Week-over-week growth %
```

---

## 8. ✏️ Latihan Mandiri

1. **Hitung capacity requirement** untuk servis yang growth 20% per bulan, current peak 500m CPU. Berapa CPU yang dibutuhkan 6 bulan lagi?
2. **Buat VPA** untuk servis `checkout-service` dengan `minAllowed.cpu: 50m` dan `maxAllowed.cpu: 2000m`.
3. **Buat alert predict_linear** yang firing ketika memory akan > 90% dalam 7 hari.
4. **Tulis capacity report** mingguan: kapan, siapa yang review, format apa.

---

**Lanjut ke Modul 04:** [04-runbook.md](./04-runbook.md) — cara membuat Runbook yang executable untuk tim on-call agar bisa merespons insiden dalam hitungan menit.
