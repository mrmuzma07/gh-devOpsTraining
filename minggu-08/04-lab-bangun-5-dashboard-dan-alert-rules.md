# Modul 04: Lab Hands-on — Bangun 5 Dashboard & 6 Alert Rules

> **Target Pembelajaran:** Hands-on membangun **5 dashboard** sesuai syllabus (Cluster, Node, Namespace, Pod, Go App) dan **6 alert rules** (Pod Restart, CPU, Memory, Disk, High Latency, Error Rate) di Alertmanager.

---

## 1. Capaian Lab

Setelah lab ini, cluster `mini-prod` Anda akan memiliki:

| Layer | Dashboard | Panel Utama |
| :--- | :--- | :--- |
| **Cluster** | Cluster Overview | Nodes ready, total pods, total namespaces, CPU/Mem gauge, disk |
| **Node** | Node Detail | CPU/Mem/Disk per node, load average |
| **Namespace** | Namespace Detail | CPU/Mem usage per namespace, pod count |
| **Pod** | Pod Detail | Status timeline, restart count, OOMKilled events |
| **Go App** | RED Method | Request rate, error rate, latency p95 |

**Alert rules (6 sesuai syllabus):**

| Alert | Severity | Tipe |
| :--- | :--- | :--- |
| Pod Restart > 5x/jam | warning | infrastructure |
| CPU > 85% | warning | infrastructure |
| Memory > 90% | warning | infrastructure |
| Disk > 85% | critical | infrastructure |
| Latency p95 > 1s | critical | application |
| Error Rate > 5% | critical | application |

Plus 2 bonus: **GoAppDown** (critical) & **PodPendingTooLong** (warning).

---

## 2. Prasyarat

```bash
# 1. Cluster, namespace, dan stack observability harus aktif
kubectl get pods -n mini-prod

# Pastikan ada: k3s, mimir, grafana, alloy, loki, tempo
```

---

## 3. Langkah 1: Deploy Alertmanager

```bash
kubectl apply -f minggu-08/manifests/02-alertmanager-config.yaml
```

**Verifikasi:**
```bash
kubectl get pods -n mini-prod -l app=alertmanager
```

**Buka UI Alertmanager** (opsional):
```bash
kubectl port-forward svc/alertmanager 9093:9093 -n mini-prod
```

Buka `http://localhost:9093` — lihat tab **Status** untuk pastikan config valid.

> ⚠️ **Penting:** Untuk lab, Slack webhook tidak perlu diisi. Alert akan tetap firing ke receiver default, hanya saja tidak benar-benar masuk ke Slack sampai webhook diisi.

---

## 4. Langkah 2: Apply Prometheus Rules

```bash
kubectl apply -f minggu-08/manifests/01-prometheus-rules.yaml
```

**Verifikasi (jika pakai kube-prometheus-stack):**
```bash
kubectl get prometheusrules -n mini-prod
```

Akan muncul:
```
NAME               AGE
mini-prod-alerts   30s
```

**Cara termudah verifikasi rule sudah aktif** — query di Grafana Explore:

```promql
# Lihat semua alert yang firing sekarang
ALERTS{alertstate="firing"}

# Lihat rule group yang sudah dimuat
count(group(prometheus_rule_evaluation_duration_seconds)) by (rule_group)
```

---

## 5. Langkah 3: Import Dashboard

Ada dua cara import dashboard. Kita pakai **cara manual** (lebih mudah debug).

### 5.1 Siapkan JSON untuk 5 Dashboard

Kita akan membuat ringkas masing-masing dashboard. Struktur tiap file ada di `minggu-08/dashboards/`.

### 5.2 Import via Grafana UI

Untuk setiap dashboard:

1. Buka Grafana → **Dashboards** → **New** → **Import**
2. Upload file JSON
3. Pilih datasource: **Prometheus** (atau Mimir jika Anda rename)
4. Klik **Import**

---

## 6. Langkah 4: Detail Tiap Dashboard

### 6.1 Dashboard 1: Cluster Overview

**File:** `dashboards/01-cluster-overview.json`

**Panel-panel utama:**

| Panel | Tipe | Query |
| :--- | :--- | :--- |
| Ready Nodes | Stat | `count(kube_node_status_condition{condition="Ready",status="true"})` |
| Total Pods | Stat | `count(kube_pod_info{namespace="mini-prod"})` |
| Namespaces | Stat | `count(kube_namespace_info)` |
| Cluster CPU % | Gauge | `100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)` |
| Cluster Memory % | Gauge | `(1 - (sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes))) * 100` |
| Pods per Namespace | Bar Chart | `count by(namespace) (kube_pod_info)` |
| Recent Restarts | Table | `topk(10, kube_pod_container_status_restarts_total)` |

**Cara baca:** Saat dashboard ini dibuka, dalam 1-2 detik Anda tahu:
- ✅ Apakah semua node ready?
- ✅ Berapa total pod berjalan?
- ✅ Cluster CPU/Mem usage overall
- ✅ Namespace mana yang paling banyak pod-nya

### 6.2 Dashboard 2: Node Detail

**File:** `dashboards/02-node-detail.json`

**Panel-panel:**

| Panel | Tipe | Query |
| :--- | :--- | :--- |
| CPU Usage per Node | Timeseries | `100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)` |
| Memory Usage per Node | Timeseries | `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100` |
| Disk Usage per Node | Timeseries | `(1 - (node_filesystem_avail_bytes{fstype!="tmpfs"} / node_filesystem_size_bytes)) * 100` |
| Load Average | Timeseries | `node_load5` |
| Network I/O | Timeseries | `rate(node_network_receive_bytes_total[5m])` |

**Variabel:** `$instance` (dropdown untuk pilih node tertentu).

### 6.3 Dashboard 3: Namespace Detail

**File:** `dashboards/03-namespace-detail.json`

| Panel | Tipe | Query |
| :--- | :--- | :--- |
| CPU by Namespace | Bar Chart | `sum by(namespace) (rate(container_cpu_usage_seconds_total{namespace!=""}[5m]))` |
| Memory by Namespace | Bar Chart | `sum by(namespace) (container_memory_working_set_bytes{namespace!=""})` |
| Pod Count by Namespace | Stat | `count by(namespace) (kube_pod_info)` |
| Network by Namespace | Timeseries | `sum by(namespace) (rate(container_network_receive_bytes_total[5m]))` |

### 6.4 Dashboard 4: Pod Detail

**File:** `dashboards/04-pod-detail.json`

| Panel | Tipe | Query |
| :--- | :--- | :--- |
| Pod Status | State Timeline | `kube_pod_status_phase{namespace="mini-prod"}` |
| Restart Count | Table | `topk(20, increase(kube_pod_container_status_restarts_total[1h]))` |
| CPU per Pod | Timeseries | `sum by(pod) (rate(container_cpu_usage_seconds_total{namespace="mini-prod"}[5m]))` |
| Memory per Pod | Timeseries | `sum by(pod) (container_memory_working_set_bytes{namespace="mini-prod"})` |
| Recent OOMKilled | Logs | `{namespace="mini-prod"} |= "OOMKilled"` |

**Variabel:** `$namespace` dan `$pod`.

### 6.5 Dashboard 5: Go App (RED Method)

**File:** `dashboards/05-go-app-red-method.json`

| Panel | Tipe | Query |
| :--- | :--- | :--- |
| **R**ate (RPS) | Stat + Timeseries | `sum(rate(http_requests_total{service="go-app"}[5m]))` |
| **E**rror Rate | Stat + Timeseries | `sum(rate(http_requests_total{service="go-app",status=~"5.."}[5m])) / sum(rate(http_requests_total{service="go-app"}[5m]))` |
| **D**uration p95 | Stat + Timeseries | `histogram_quantile(0.95, sum by(le) (rate(http_request_duration_seconds_bucket{service="go-app"}[5m])))` |
| RPS per Endpoint | Timeseries | `sum by(path) (rate(http_requests_total{service="go-app"}[5m]))` |
| Latency p50/p95/p99 | Timeseries | `histogram_quantile(0.50, ...)` / `0.95` / `0.99` |
| Errors by Status | Bar Chart | `sum by(status) (rate(http_requests_total{service="go-app",status=~"5.."}[5m]))` |

**Panel header info:**
- Threshold line latency: 1000ms (warna merah — alert threshold)
- Threshold line error rate: 5% (warna merah)

---

## 7. Langkah 5: Generate Metrics Agar Dashboard Tidak Kosong

Pastikan Go App dari Minggu 7 mengirim metrik. Jika belum ada exporter Prometheus, tambahkan:

```bash
# Install Prometheus Go client (sudah ada di kebanyakan framework)
# Pastikan Go App expose endpoint /metrics
curl http://localhost:8080/metrics
```

Generate beberapa request untuk mengisi data:

```bash
kubectl port-forward svc/go-app-service 8080:80 -n mini-prod

# 100 request untuk generate RPS
for i in $(seq 1 100); do
  curl -s "http://localhost:8080/order?product=$i" > /dev/null
done
```

Sekarang buka dashboard `05-go-app-red-method.json` — panel RPS, Error Rate, Latency akan terisi.

---

## 8. Langkah 6: Verifikasi Alert Rule Active

### 8.1 Cek di Alertmanager UI

```bash
kubectl port-forward svc/alertmanager 9093:9093 -n mini-prod
```

Buka `http://localhost:9093`:
- Tab **Status** → lihat "Config" valid
- Tab **Alerts** → saat ini belum ada (normal — alert firing hanya saat threshold dilanggar)

### 8.2 Test Alert dengan Trigger Manual

**Test 1: Trigger alert GoAppDown** (kebetulan, jangan dilakukan di production!)

```bash
# Scale down Go App ke 0 replicas
kubectl scale deployment/go-app -n mini-prod --replicas=0

# Tunggu 2 menit (for: 2m di rule)
# Cek alert:
# kubectl exec ... alertmanager -- amtool alert query

# Scale up kembali
kubectl scale deployment/go-app -n mini-prod --replicas=1
```

**Test 2: Trigger Pod Restart alert**

```bash
# Hapus Pod Go App berkali-kali
for i in 1 2 3 4 5 6; do
  kubectl delete pod -l app=go-app -n mini-prod --wait=false
  sleep 5
done

# Tunggu 10+ menit (for: 10m)
# Alert "PodRestartingFrequently" akan firing
```

---

## 9. Langkah 7: Drilldown Navigation

Buat tiap dashboard saling link supaya engineer bisa zoom-in/out:

```json
{
  "links": [
    {
      "title": "Cluster Overview",
      "url": "/d/minggu-08-cluster-overview",
      "type": "link"
    },
    {
      "title": "Node Detail",
      "url": "/d/minggu-08-node-detail",
      "type": "link"
    },
    {
      "title": "Namespace Detail",
      "url": "/d/minggu-08-namespace-detail?var-namespace=${namespace}",
      "type": "link"
    },
    {
      "title": "Go App RED",
      "url": "/d/minggu-08-go-app-red",
      "type": "link"
    }
  ]
}
```

Sekarang dari Cluster Overview → klik namespace → masuk ke Namespace Detail dengan filter otomatis.

---

## 10. Checklist Capaian Lab

```bash
✅ Alertmanager Pod Running
✅ PrometheusRule applied (8 rules aktif)
✅ Dashboard Cluster Overview imported
✅ Dashboard Node Detail imported
✅ Dashboard Namespace Detail imported
✅ Dashboard Pod Detail imported
✅ Dashboard Go App (RED) imported
✅ Go App mengirim metrics (RPS, error, latency)
✅ Drilldown links antar dashboard berfungsi
```

---

## 11. Troubleshooting

| Gejala | Solusi |
| :--- | :--- |
| Rule tidak aktif | Cek label `prometheus: mimir` match dengan selector Prometheus Anda |
| Dashboard panel "No data" | Cek datasource UID di panel settings, pastikan match dengan Prometheus/Mimir UID |
| Alert firing tapi tidak masuk Slack | Pastikan webhook URL sudah diganti di `alertmanager.yaml` |
| Alert firing terus-menerus | Tune `for` duration atau threshold |
| Alertmanager config error | Cek log: `kubectl logs -l app=alertmanager -n mini-prod` |

---

## 12. Rangkuman

Di modul ini Anda sudah:
- ✅ Deploy **Alertmanager** dengan routing critical/warning
- ✅ Apply **6 alert rules** sesuai syllabus (+ 2 bonus)
- ✅ Membangun **5 dashboard** sesuai template Modul 02
- ✅ Memahami drilldown navigation antar dashboard
- ✅ Generate metrics agar dashboard tidak kosong
- ✅ Test alert manual untuk verifikasi

Lanjut ke **Modul 05** untuk simulasi **alert storm** dan latihan tuning threshold.