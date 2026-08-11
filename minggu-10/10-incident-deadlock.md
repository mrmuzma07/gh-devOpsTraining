# Incident #10 — Database Deadlocks (PostgreSQL Error 40P01)

> **Satu kalimat:** Ketika dua atau lebih transaksi database berjalan secara bersamaan dan **saling mengunci (lock) resource yang dibutuhkan oleh transaksi lawan secara bersilang**, PostgreSQL akan membatalkan salah satu transaksi dengan error `deadlock detected (SQLSTATE 40P01)`.

Insiden **Deadlock** adalah salah satu insiden tingkat lanjut yang paling menantang karena terjadi secara **non-deterministic (tergantung timing / race condition)** saat trafik tinggi. Di level Kubernetes, aplikasi memuntahkan `HTTP 500` secara acak pada endpoint bertransaksi tinggi (seperti transfer uang, reservasi tiket, atau checkout stok).

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Memahami mekanisme terjadinya **Circular Deadlock** pada Database Relasional (Row-Level Locking).
2. Membaca dan menganalisis PostgreSQL Deadlock Graph & Error Logs (`SQLSTATE 40P01`).
3. Mendiagnosis lock contention menggunakan query `pg_locks` dan `pg_stat_activity`.
4. Mengimplementasikan teknik **Consistent Locking Order** dan **Exponential Backoff Retry Strategy** di kode aplikasi.
5. Memonitor metrics PostgreSQL Deadlock Rate (`pg_stat_database_deadlocks`).

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari Sisi Application Logs (Loki)
Aplikasi memuntahkan exception database saat mengeksekusi transaksi:
```text
2026-08-11T03:45:10.102Z [ERROR] [order-service] Failed to complete transfer transaction: 
  pq: deadlock detected 
  DETAIL: Process 19283 waits for ShareLock on transaction 89123; blocked by process 19284.
          Process 19284 waits for ShareLock on transaction 89122; blocked by process 19283.
  HINT: See server log for query details.
  LOCATION: DeadlockCheck, lock.c:3819
  SQLSTATE: 40P01
```

### 1.2 Dari Sisi Grafana Mimir & Metrics
- Metric `pg_stat_database_deadlocks` membengkak naik:
```promql
# Rate of Deadlocks per second
rate(pg_stat_database_deadlocks{datname="appdb"}[5m]) > 0
# Hasil: 4.2 deadlocks/sec selama periode flash sale!
```
- Spike pada HTTP 500 Error Rate khusus pada endpoint pembayaran/transfer.

---

## 2. 🧠 Mekanisme Terjadinya Deadlock (The Circular Wait)

Bayangkan dua transaksi transfer uang terjadi pada detik yang persis sama:
- **Transaksi 1 (T1):** User A mengirim uang ke User B.
- **Transaksi 2 (T2):** User B mengirim uang ke User A.

```text
[ Transaksi 1 (T1) ]                          [ Transaksi 2 (T2) ]
        │                                              │
        ▼                                              ▼
1. LOCK Row User A (SUCCESS)                  1. LOCK Row User B (SUCCESS)
        │                                              │
        ▼                                              ▼
2. Mencoba LOCK Row User B                    2. Mencoba LOCK Row User A
   └─► WAITING... (B locked by T2)               └─► WAITING... (A locked by T1)
        │                                              │
        └────────────────── CIRCULAR WAIT ─────────────┘
                                │
                                ▼
               [ PostgreSQL Deadlock Detector ]
              Memilih 1 transaksi untuk di-ABORT:
              - T1: SUCCESS (melanjutkan commit)
              - T2: ABORTED dengan Error 40P01 (Rolled Back!)
```

---

## 3. 🔍 Investigasi Step-by-Step

### Step 1 — Periksa Deadlock Log di Loki / PostgreSQL
PostgreSQL secara otomatis mencatat detail pertikaian lock di log file ketika deadlock terjadi:

```bash
$ kubectl logs postgres-0 -n database | grep -A 10 "deadlock detected"
```

Sample Output Log PostgreSQL:
```text
2026-08-11 03:45:10 UTC [19283] ERROR:  deadlock detected
2026-08-11 03:45:10 UTC [19283] DETAIL:  Process 19283 waits for ExclusiveLock on tuple (0,12) of relation 16384 of database 16385; blocked by process 19284.
Process 19284 waits for ExclusiveLock on tuple (0,8) of relation 16384 of database 16385; blocked by process 19283.
Process 19283: UPDATE accounts SET balance = balance - 100 WHERE id = 10;
Process 19284: UPDATE accounts SET balance = balance - 50 WHERE id = 5;
```

> **Analisis:**
> - Process 19283 sedang memegang lock untuk `id = 5` dan ingin mengunci `id = 10`.
> - Process 19284 sedang memegang lock untuk `id = 10` dan ingin mengunci `id = 5`.

---

### Step 2 — Periksa Transaksi Hang / Waiting Locks (`pg_locks`)
Jika deadlock belum ter-cancel dan transaksi gantung (lock timeout):

```sql
SELECT 
    blocked_locks.pid     AS blocked_pid,
    blocked_activity.usename  AS blocked_user,
    blocking_locks.pid    AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query    AS blocked_statement
FROM  pg_catalog.pg_locks         blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks         blocking_locks 
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;
```

---

## 4. 🛠️ Penanganan & Solusi Perbaikan (Code Level Fix)

Deadlock **TIDAK BISA** diperbaiki hanya dengan mengubah setting infrastruktur Kubernetes / Postgres. Kunci perbaikan ada di **Logika Kode Aplikasi**.

### Solusi 1 — Consistent Locking Order (Urutkan Lock berdasarkan ID)
Selalu urutkan Resource ID yang akan di-update secara konsisten (misal: urutkan ID dari terkecil ke terbesar).

```go
// KODE BAGUS: Lock dieksekusi berdasarkan urutan ID terkecil dahulu
func TransferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Tentukan urutan locking agar selalu konsisten!
    firstID, secondID := fromID, toID
    if firstID > secondID {
        firstID, secondID = toID, fromID
    }

    // 1. Lock ID yang lebih kecil dahulu
    _, err = tx.Exec("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", firstID)
    if err != nil { return err }

    // 2. Lock ID yang lebih besar kemudian
    _, err = tx.Exec("SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", secondID)
    if err != nil { return err }

    // 3. Lakukan proses Update
    _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
    _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)

    return tx.Commit()
}
```

---

### Solusi 2 — Application Retry Mechanism (Exponential Backoff)
Karena Deadlock bersifat sementara (transaksi lawan akan selesai dalam beberapa milidetik), aplikasi **HARUS** menangani error `40P01` dengan melakukan **Retry**:

```go
// Driver Level Retry untuk Error 40P01
func ExecuteWithRetry(fn func() error) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err != nil && isDeadlockError(err) { // SQLSTATE 40P01
            time.Sleep(time.Duration(100*(i+1)) * time.Millisecond) // Exponential Backoff
            continue
        }
        return err
    }
    return fmt.Errorf("transaction failed after max retries")
}
```

---

## 5. 🛡️ Strategi Pencegahan Jangka Panjang

### 5.1 Lower `deadlock_timeout` di PostgreSQL
Secara default, PostgreSQL menunggu `1s` sebelum mengaktifkan Deadlock Detection. Turunkan menjadi `200ms` agar deadlock dideteksi dan dibatalkan lebih cepat tanpa menahan koneksi pool terlalu lama.

```sql
ALTER DATABASE appdb SET deadlock_timeout = '200ms';
```

### 5.2 Singkatkan Durasi Transaksi (Keep Transactions Short)
- Jangan pernah melakukan panggilan HTTP API External, I/O File, atau enkripsi berat di **DALAM** blok transaksi database (`BEGIN ... COMMIT`).
- Lakukan persiapkan data terlebih dahulu, baru buka transaksi DB di akhir proses.

### 5.3 Prometheus Alerting untuk Deadlock Detection
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: deadlock-alerts
  namespace: monitoring
spec:
  groups:
  - name: deadlock-rules
    rules:
    - alert: PostgreSQLDeadlockSpike
      expr: rate(pg_stat_database_deadlocks{datname="appdb"}[5m]) > 0.5
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "Spike Deadlock Terdeteksi di Database appdb! Rate: {{ $value | printf \"%.2f\" }}/sec"
```

---

## 6. 🧪 Lab Hands-On: Reproduksi Deadlock

### Step 1 — Jalankan 2 Session Terminal `psql` Bersamaan

Terminal 1 (Transaksi 1):
```sql
BEGIN;
UPDATE accounts SET balance = balance - 10 WHERE id = 1;
-- Jangan COMMIT dulu!
```

Terminal 2 (Transaksi 2):
```sql
BEGIN;
UPDATE accounts SET balance = balance - 20 WHERE id = 2;
-- Jangan COMMIT dulu!
```

### Step 2 — Eksekusi Circular Cross Update
Terminal 1:
```sql
UPDATE accounts SET balance = balance + 10 WHERE id = 2;
-- Terminal 1 akan HANG (Waiting Lock dari Terminal 2)
```

Terminal 2:
```sql
UPDATE accounts SET balance = balance + 20 WHERE id = 1;
-- BOOM! PostgreSQL langsung mendeteksi Deadlock di Terminal 2!
```

Output Terminal 2:
```text
ERROR:  deadlock detected
DETAIL:  Process 12903 waits for ShareLock on transaction 8812; blocked by process 12902.
...
SQLSTATE: 40P01
```

---

## 7. 📋 Cheat Sheet Deadlock Remediation

```text
DIAGNOSIS STEP                   | TOOL / COMMAND                               | TINDAKAN PERBAIKAN
---------------------------------|----------------------------------------------|----------------------------------------------------------
1. Detect Deadlock Error         | Application Logs (SQLSTATE 40P01)           | Konfirmasi terjadi race condition transaksi
2. Check Deadlock Rate           | PromQL: rate(pg_stat_database_deadlocks[5m]) | Ukur frekuensi kejadian insiden saat peak traffic
3. Inspect Lock Graph            | Query pg_locks & pg_stat_activity            | Identifikasi tabel & row yang saling bertikai
4. Application Fix               | Consistent Lock Order (sort IDs)             | Wajibkan seluruh query mengunci resource dengan urutan sama
5. Resilience Fix                | Exponential Backoff Retry                    | Tangani error 40P01 dengan retry otomatis 3x
```

---

**Lanjut ke Modul Penutup:** [README.md](./README.md) — Rangkuman seluruh 9 simulasi insiden kompleks Week 10 & panduan ujian integrasi SRE.
