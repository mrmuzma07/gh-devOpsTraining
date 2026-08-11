# Incident #9 — Slow Database Queries & Missing Indexes

> **Satu kalimat:** Ketika database PostgreSQL/MySQL kehabisan CPU dan disk I/O, **request aplikasi akan mengalami antrean (lock waiting & connection pool exhaustion)** yang disebabkan oleh query berat tanpa Indeks (`Sequential Scan`) atau transaksi gantung.

Berbeda dengan Insiden #8 (N+1 Query yang dipicu oleh jumlah query kecil yang sangat banyak), Insiden #9 dipicu oleh **Satu atau Beberapa Query Tunggal yang Sangat Berat** yang memaksa PostgreSQL membaca jutaan baris data dari disk (`Sequential Scan` pada tabel raksasa).

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Mendiagnosis query PostgreSQL yang lambat menggunakan `pg_stat_statements` dan `EXPLAIN ANALYZE`.
2. Memahami perbedaan antara **Sequential Scan** (Bad) vs **Index Scan / Index Only Scan** (Good).
3. Mengidentifikasi masalah **PostgreSQL Connection Pool Exhaustion** (`max_connections` reached).
4. Melakukan pembuatan indeks secara aman di Production menggunakan `CREATE INDEX CONCURRENTLY`.
5. Memonitor metrics PostgreSQL menggunakan `postgres-exporter` dan Grafana Mimir.

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari Sisi Grafana Mimir & Prometheus
- Metric CPU Utilization Pod PostgreSQL menyentuh **100% (CPU Throttling / Resource Limit Reached)**.
- Metric Active Connections menyentuh batas maksimum `max_connections` (misal: 100/100 connections in use).
- Metric PromQL:
```promql
# Persentase Query yang melakukan Sequential Scan dibanding Index Scan
sum(rate(pg_stat_user_tables_seq_scan[5m])) 
/ (sum(rate(pg_stat_user_tables_seq_scan[5m])) + sum(rate(pg_stat_user_tables_idx_scan[5m]))) * 100
# Hasil: 88% Sequential Scan! (Sangat Bahaya)
```

### 1.2 Dari Sisi Log PostgreSQL (Loki)
```text
2026-08-11T03:30:12.102Z [POSTGRES] LOG: duration: 12450.123 ms statement: 
    SELECT * FROM transactions WHERE status = 'PENDING' AND created_at >= '2025-01-01' ORDER BY amount DESC;
2026-08-11T03:30:14.400Z [POSTGRES] WARNING: remaining connection slots are reserved for non-replication superuser connections
```

### 1.3 Dari Sisi Aplikasi
- Error `500 Internal Server Error` dengan pesan:
  `pq: sorry, too many clients already` atau `driver: bad connection / connection pool timeout`.

---

## 2. 🔍 Investigasi Step-by-Step

### Step 1 — Periksa Query Aktif di PostgreSQL (`pg_stat_activity`)
Exec ke dalam Pod Database PostgreSQL:
```bash
$ kubectl exec -it postgres-0 -n database -- psql -U postgres -d appdb
```
Jalankan query analisis real-time untuk melihat query apa yang sedang memakan waktu paling lama:

```sql
SELECT 
    pid, 
    now() - query_start AS duration, 
    state, 
    query 
FROM pg_stat_activity 
WHERE state != 'idle' 
ORDER BY duration DESC 
LIMIT 5;
```

Sample Output:
```text
  pid  |    duration     | state  |                                                    query                                                    
-------+-----------------+--------+-------------------------------------------------------------------------------------------------------------
 18291 | 00:00:14.281923 | active | SELECT * FROM transactions WHERE status = 'PENDING' AND created_at >= '2025-01-01' ORDER BY amount DESC;
 18302 | 00:00:12.102911 | active | SELECT * FROM transactions WHERE status = 'PENDING' AND created_at >= '2025-01-01' ORDER BY amount DESC;
```

> **Temuan 1:** Terdapat multiple query identik yang berjalan selama **12-14 detik**!

---

### Step 2 — Analisis Query Plan Menggunakan `EXPLAIN ANALYZE`
Jalankan perintah `EXPLAIN ANALYZE` untuk query tersebut:

```sql
EXPLAIN ANALYZE 
SELECT * FROM transactions WHERE status = 'PENDING' AND created_at >= '2025-01-01' ORDER BY amount DESC;
```

Sample Output:
```text
QUERY PLAN
-----------------------------------------------------------------------------------------------------------------------
 Sort  (cost=18520.00..18770.00 rows=100000 width=128) (actual time=12100.12..12350.45 rows=95000 loops=1)
   Sort Key: amount DESC
   Sort Method: external merge Disk: 18500kB
   ->  Seq Scan on transactions  (cost=0.00..12500.00 rows=100000 width=128) (actual time=0.05..8900.12 rows=95000 loops=1)
         Filter: ((status = 'PENDING'::text) AND (created_at >= '2025-01-01 00:00:00'::timestamp))
         Rows Removed by Filter: 4050000
 Execution Time: 12410.50 ms
```

### Breakdown Analisis Query Plan:
1. `Seq Scan on transactions`: Postgres membaca seluruh **4.150.000 baris data dari disk**.
2. `Rows Removed by Filter: 4050000`: Postgres membuang 4 juta baris karena tidak cocok.
3. `Execution Time: 12410 ms`: Query membutuhkan waktu **12.4 detik**!

---

## 3. 🛠️ Penanganan & Mitigasi (Step-by-Step Fix)

### Tahap 1 — Emergency Fix (Kill Long-Running Slow Queries)
Jika database dalam kondisi hang total akibat query ini:

```sql
-- Kill query spesifik berdasarkan PID yang berjalan > 10 detik
SELECT pg_cancel_backend(18291); -- Graceful cancel
SELECT pg_terminate_backend(18291); -- Force terminate connection
```

---

### Tahap 2 — Permanent Fix (Buat Indeks yang Tepat)

Buat Composite Index pada kolom `status` dan `created_at` serta `amount`:

> ⚠️ **ATURAN PROD KUBE/SRE:** Jangan pernah jalankan `CREATE INDEX` biasa di tabel besar pada Production! Karena perintah tersebut akan **LOCK TABEL (Write Lock)** dan menyebabkan aplikasi crash!
> Gunakan perintah **`CREATE INDEX CONCURRENTLY`**.

```sql
-- Pembuatan Indeks secara Non-Blocking di Production
CREATE INDEX CONCURRENTLY idx_transactions_status_created 
ON transactions (status, created_at, amount DESC);
```

---

### Step 3 — Verifikasi Ulang `EXPLAIN ANALYZE` Setelah Indeks

Jalankan kembali query planner:
```sql
EXPLAIN ANALYZE 
SELECT * FROM transactions WHERE status = 'PENDING' AND created_at >= '2025-01-01' ORDER BY amount DESC;
```

Sample Output Setelah Fix:
```text
QUERY PLAN
-----------------------------------------------------------------------------------------------------------------------
 Index Scan using idx_transactions_status_created on transactions  (cost=0.42..120.50 rows=95000 width=128) (actual time=0.04..12.30 rows=95000 loops=1)
 Execution Time: 14.20 ms  <-- TURUN DARI 12,400ms MENJADI 14ms! (900x LEBIH CEPT!)
```

---

## 4. 🛡️ Strategi Pencegahan Jangka Panjang

### 4.1 Aktifkan `pg_stat_statements` Extension
`pg_stat_statements` mencatat seluruh statistik query secara berkala di PostgreSQL:

```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Lihat 5 Query Paling Banyak Memakan Total Waktu Execution
SELECT 
    calls, 
    round(total_exec_time::numeric, 2) as total_ms, 
    round(mean_exec_time::numeric, 2) as avg_ms, 
    query 
FROM pg_stat_statements 
ORDER BY total_exec_time DESC 
LIMIT 5;
```

### 4.2 Batasi Timeout Query di PostgreSQL (`statement_timeout`)
Mencegah query "liar" berjalan selamanya dan menghabisi CPU:

```sql
-- Set maksimal query berjalan adalah 5 detik (5000ms)
ALTER DATABASE appdb SET statement_timeout = '5000ms';
```

### 4.3 Pasang PgBouncer (Connection Pooler)
Gunakan `PgBouncer` sebagai sidecar atau Deployment terpisah di depan PostgreSQL untuk mengelola ribuan koneksi dari Kubernetes Pods tanpa menghabisi `max_connections` milik PostgreSQL.

---

## 5. 🧪 Lab Hands-On: Slow Query & Indexing

### Step 1 — Buat Tabel & Data Simulasi
File: `minggu-10/manifests/09-postgres-slow-query-lab.yaml`
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pg-init-script
  namespace: default
data:
  init.sql: |
    CREATE TABLE IF NOT EXISTS logs_data (
        id SERIAL PRIMARY KEY,
        log_level VARCHAR(10),
        message TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    -- Insert 100.000 dummy rows
    INSERT INTO logs_data (log_level, message)
    SELECT 
        (ARRAY['INFO', 'WARN', 'ERROR'])[floor(random() * 3 + 1)],
        'Simulated log message number ' || g.i
    FROM generate_series(1, 100000) g(i);
```

### Step 2 — Simulasikan Query Tanpa Index vs Dengan Index
```bash
# 1. Jalankan Query Tanpa Index (Seq Scan)
$ kubectl exec -it deployment/postgres -- psql -U postgres -c "EXPLAIN ANALYZE SELECT * FROM logs_data WHERE log_level = 'ERROR';"

# 2. Tambahkan Index
$ kubectl exec -it deployment/postgres -- psql -U postgres -c "CREATE INDEX CONCURRENTLY idx_logs_level ON logs_data(log_level);"

# 3. Jalankan Query Setelah Index (Index Scan)
$ kubectl exec -it deployment/postgres -- psql -U postgres -c "EXPLAIN ANALYZE SELECT * FROM logs_data WHERE log_level = 'ERROR';"
```

---

## 6. 📋 Cheat Sheet Slow Database Tuning

```text
PROBLEM                          | METRIC / TOOL                           | SOLUSI / COMMAND
---------------------------------|-----------------------------------------|----------------------------------------------------------
Sequential Scan High             | PromQL: pg_stat_user_tables_seq_scan    | CREATE INDEX CONCURRENTLY idx_name ON table(col);
Max Connections Reached          | Metric: pg_stat_database_numbackends   | Implementasi PgBouncer / Scale App Connection Pool
Long Running Query (> 10s)       | Query: pg_stat_activity                 | SELECT pg_terminate_backend(pid);
Unbounded Query Execution        | Config: statement_timeout               | SET statement_timeout = '5000ms';
Query Lock Waiting               | Query: pg_locks                         | Periksa transaksi gantung / deadlock (Insiden #10)
```

---

**Lanjut ke insiden berikutnya:** [10-incident-deadlock.md](./10-incident-deadlock.md) — menangani insiden deadlock database (PostgreSQL Error `40P01`) akibat persaingan urutan update baris data secara bersamaan.
