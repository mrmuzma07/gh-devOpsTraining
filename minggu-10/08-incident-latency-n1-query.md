# Incident #8 — High Latency & N+1 Database Query Pattern

> **Satu kalimat:** Ketika endpoint API menjadi sangat lambat (latency > 5-10 detik) saat jumlah data meningkat, penyebab yang sangat sering ditemui adalah **Anti-Pattern N+1 Query**, di mana aplikasi mengeksekusi 1 query awal untuk mengambil daftar item, diikuti oleh N query terpisah secara berulang (loop) untuk setiap detail item tersebut.

Insiden latency tidak selalu disebabkan oleh kekurangan resource CPU/Memory atau kerusakan infrastruktur Kubernetes. Seringkali infrastruktur Kubernetes berjalan sempurna, tetapi **kode aplikasi atau penggunaan ORM (Object-Relational Mapping) yang buruk** menciptakan _database round-trip overhead_ yang masif.

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Memahami mekanisme terjadinya **N+1 Query Problem** pada arsitektur Microservice / ORM.
2. Membaca **Distributed Trace Waterfall Chart di Grafana Tempo** untuk mengidentifikasi pola n-spans berulang.
3. Menganalisis korelasi antara Latency Spike (Mimir) dengan Log Query Database (Loki).
4. Melakukan refactoring query dari N+1 (`SELECT * FROM items WHERE id = ?`) menjadi Batch Query (`JOIN` atau `IN (...)`).
5. Memasang Alerting Latency SLA/SLO (p95 / p99 duration).

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari Sisi User Experience / API Client
- Panggilan ke `/api/v1/orders` membutuhkan waktu 8.5 detik untuk merespons 100 data transaksi.
- Aplikasi Frontend (React/Mobile) mengalami "spinners/loaders" yang lama dan memicu timeout di sisi Ingress/Gateway.

### 1.2 Dari Sisi Grafana Mimir (Metrics)
- Metric Histogram Latency p95/p99 melonjak tinggi saat jumlah item di database bertambah:
```promql
# Rate Request Latency 95th Percentile
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="order-service"}[5m])) by (le))
# Hasil: 8.42s (Padahal baseline SLA < 200ms!)
```
- Penggunaan CPU Database Server (Postgres) naik bertahap, sementara CPU Application Pod tetap stabil rendah (menunggu I/O network DB).

### 1.3 Dari Sisi Grafana Loki (Logs)
Log database dibanjiri puluhan ribu query berurutan dalam waktu kurang dari 1 detik:
```text
2026-08-11T03:10:01.001Z [DB-QUERY] SELECT * FROM users WHERE id = 1; (duration: 2ms)
2026-08-11T03:10:01.004Z [DB-QUERY] SELECT * FROM users WHERE id = 2; (duration: 2ms)
2026-08-11T03:10:01.007Z [DB-QUERY] SELECT * FROM users WHERE id = 3; (duration: 2ms)
... (100x berulang dalam 1 HTTP request)
```

---

## 2. 🧠 Visualisasi Pola N+1 di Tempo Tracing

Gunakan **Grafana Tempo** untuk membuktikan insiden ini secara visual.

### Trace Waterfall Normal (Optimized JOIN / Batch):
```text
[HTTP GET /api/v1/orders] ======================================> 45ms
  ├── [sql.query] SELECT orders JOIN users JOIN items... ========> 38ms (1 Single Query)
  └── [json.marshal] Render Response JSON =======================> 4ms
```

### Trace Waterfall Insiden N+1 Query:
```text
[HTTP GET /api/v1/orders] ==========================================================================> 8,450ms
  ├── [sql.query] SELECT * FROM orders; =====================> 10ms (Query 1)
  ├── [sql.query] SELECT * FROM users WHERE id = 1; ========> 80ms (Query 2 - Roundtrip 1)
  ├── [sql.query] SELECT * FROM users WHERE id = 2; ========> 82ms (Query 3 - Roundtrip 2)
  ├── [sql.query] SELECT * FROM users WHERE id = 3; ========> 81ms (Query 4 - Roundtrip 3)
  │   ... (97 Spans serupa berderet kebawah)
  └── [sql.query] SELECT * FROM users WHERE id = 100; ======> 79ms (Query 101 - Roundtrip 100)
```

> **Diagnosis Utama Tempo:** Terlihat "tangga waterfall" panjang yang berisi puluhan hingga ratusan span database query kecil dengan nama dan struktur yang identik.

---

## 3. 🔍 Investigasi Step-by-Step

### Step 1 — Temukan Trace ID dengan Latency Tinggi dari Mimir/Loki
Cari log request di Loki yang durasinya > 5 detik untuk mendapatkan `trace_id`:
```logql
{app="order-service"} |= "GET /api/v1/orders" | json | duration > 5s
```
Sample Output:
```json
{
  "timestamp": "2026-08-11T03:10:00Z",
  "method": "GET",
  "path": "/api/v1/orders",
  "duration_ms": 8450,
  "trace_id": "a8f92b1c34891e01"
}
```

### Step 2 — Query Trace ID di Grafana Tempo
1. Buka Grafana -> Explore -> Pilih Datasource **Tempo**.
2. Masukkan Query Trace ID: `a8f92b1c34891e01`.
3. Periksa panel **Trace Details**:
   - Total Spans: `102 spans`.
   - Service Name: `order-service` -> `postgres-db`.
   - Span Attributes:
     - `db.system`: `postgresql`
     - `db.statement`: `SELECT * FROM users WHERE id = $1`

---

## 4. 🎯 Root Cause & Perbaikan Kode (Code Fix)

### Penyebab Utama (Kode Bermasalah):
Aplikasi (Go / Node.js ORM) mengambil data transaksi, lalu melakukan `looping` mengeksekusi query user satu per satu:

```go
// KODE BURUK (N+1 Query Issue)
func GetOrders(w http.ResponseWriter, r *http.Request) {
    var orders []Order
    db.Raw("SELECT * FROM orders LIMIT 100").Scan(&orders) // 1 Query

    for i := range orders {
        var user User
        // N Query! Terjadi 100x network round-trip ke Postgres!
        db.Raw("SELECT * FROM users WHERE id = ?", orders[i].UserID).Scan(&user) 
        orders[i].User = user
    }

    json.NewEncoder(w).Encode(orders)
}
```

---

### Solusi Perbaikan Kode (Refactored Code):

Gunakan **SQL JOIN** atau **Preload / Eager Loading** (`IN` clause) untuk mengubah 101 query menjadi 1 atau 2 query:

```go
// KODE BAGUS (Optimized Single JOIN Query)
func GetOrders(w http.ResponseWriter, r *http.Request) {
    type OrderResponse struct {
        OrderID   int    `json:"order_id"`
        Amount    float64 `json:"amount"`
        UserName  string `json:"user_name"`
        UserEmail string `json:"user_email"`
    }

    var results []OrderResponse
    
    // Cuma 1 Query untuk seluruh data!
    query := `
        SELECT o.id as order_id, o.amount, u.name as user_name, u.email as user_email
        FROM orders o
        INNER JOIN users u ON o.user_id = u.id
        LIMIT 100
    `
    db.Raw(query).Scan(&results)

    json.NewEncoder(w).Encode(results)
}
```

---

## 5. 🛡️ Langkah Pencegahan Jangka Panjang

### 5.1 OpenTelemetry Auto-Instrumentation
Pastikan SDK OpenTelemetry di aplikasi dikonfigurasi untuk mencatat **DB Spans & Query Statement**, sehingga ketika ada N+1 di staging environment, tim QA / Developer langsung mengetahuinya sebelum rilis ke Production.

### 5.2 Linting & Query Guard di CI/CD Pipeline
- Gunakan tool static analysis seperti `go-vet` / ORM strict mode (misal `gorm` mode `DryRun` atau `gorm.ErrEmptySlice`).
- Pasang Test Integration yang menguji jumlah query (`AssertQueryCount(t, 2)`).

### 5.3 Prometheus Latency Alert (SLO Warning)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: latency-slo-alerts
  namespace: monitoring
spec:
  groups:
  - name: latency-rules
    rules:
    - alert: APIHighLatencyP95
      expr: |
        histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="order-service"}[5m])) by (le)) > 1.0
      for: 3m
      labels:
        severity: warning
      annotations:
        summary: "Order Service p95 Latency melebihi SLO 1 Detik! Current: {{ $value | printf \"%.2f\" }}s"
```

---

## 6. 🧪 Lab Hands-On: Visualisasi N+1 di Tempo

### Step 1 — Deploy Simulation App
File: `minggu-10/manifests/08-n1-app-lab.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: n1-demo-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: n1-demo
  template:
    metadata:
      labels:
        app: n1-demo
    spec:
      containers:
      - name: app
        image: alpine:3.19
        command: ["/bin/sh", "-c"]
        args:
        - |
          echo "Simulasi N+1 Tracing App Running..."
          sleep 3600
```

### Step 2 — Verifikasi dengan Trace Viewer
- Buka Grafana Tempo Dashboard.
- Cari trace dengan span tag `db.statement` berulang.
- Konfirmasi penurunannya setelah patch query diterapkan.

---

## 7. 📋 Cheat Sheet Identification N+1 Query

```text
DIAGNOSIS PATH              | TOOL / OBSERVABILITY               | CIRI UTAMA / KUNCI
----------------------------|------------------------------------|--------------------------------------------
1. Alert Metrics Triggered  | Grafana Mimir (PromQL)             | p95/p99 HTTP latency spike (> 2-5 detik)
2. Trace Analysis           | Grafana Tempo                      | "Waterfall Staircase" dengan 50+ span DB identik
3. Log Correlation          | Grafana Loki                       | Puluhan SQL SELECT query per millisecond
4. Database Impact          | Postgres pg_stat_activity          | `calls` count naik pesat, query time singkat
5. Solusi                   | Application Code Refactoring       | Gunakan `JOIN`, `IN(...)`, atau ORM Eager Loading
```

---

**Lanjut ke insiden berikutnya:** [09-incident-slow-database.md](./09-incident-slow-database.md) — menangani degradasi database akibat query berat tanpa Index (`Seq Scan`) dan CPU Throttling pada Database Pod.
