# Modul 03: Instrumentasi JSON Structured Logging pada Go App

> **Target Pembelajaran:** Memperbarui aplikasi Go (Minggu 2) untuk menggunakan **JSON Structured Logger** dengan level log (`INFO`, `WARN`, `ERROR`), agar log-nya dapat langsung di-filter secara presisi di Grafana Loki.

---

## 1. Mengapa Aplikasi Anda Harus Logging?

Log adalah **sumber kebenaran satu-satunya** ketika terjadi insiden:
- Berapa kali request error dalam 5 menit terakhir?
- Apakah user ID tertentu mengalami error login?
- Kapan terakhir kali Pod Nginx menerima koneksi yang ditolak?

Tanpa logging yang baik, troubleshooting pada lingkungan produksi hanya dilakukan secara spekulatif (*asal tebak*).

---

## 2. Best Practice Logging di Level Kode Aplikasi

Prinsip-prinsip utama yang harus dipatuhi:
1. **Gunakan Library Logger Standar:** Jangan gunakan `fmt.Println()` untuk kebutuhan production. Gunakan library populer seperti **Logrus** atau **Zap** (`uber-go/zap`).
2. **Selalu Pakai Level Log:** Pisahkan log menjadi `INFO`, `WARN`, `ERROR`, `DEBUG`.
3. **Format Output JSON:** Setiap log harus berupa JSON string untuk muxing otomatis di Loki.
4. **Tulis ke `stdout`:** Jangan pernah menulis log ke file (`.log`) di dalam container.

---

## 3. Contoh Sintaks JSON Logger (Go)

Berikut adalah snippet kode Go App yang telah dimodifikasi (untuk `minggu-06/app/main.go`) menggunakan library `slog` (sudah built-in di Go 1.21+):

```go
package main

import (
    "log/slog"
    "os"
    "net/http"
)

func main() {
    // 1. Buat JSON logger yang menulis ke stdout
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // 2. Log pesan INFO ketika server dinyalakan
    logger.Info("server starting", "port", 8080)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        logger.Info("incoming request", "method", r.Method, "path", r.URL.Path)

        // 3. Simulasi error (akan dibahas di Lab Incident)
        if r.URL.Path == "/panic" {
            logger.Error("manual panic triggered", "client_ip", r.RemoteAddr)
            panic("Simulasi Panic Logs!")
        }

        w.Write([]byte("Hello"))
    })

    http.ListenAndServe(":8080", nil)
}
```

### Contoh Output:
```json
{"time":"2026-08-10T12:00:00Z","level":"INFO","msg":"server starting","port":8080}
{"time":"2026-08-10T12:00:05Z","level":"INFO","msg":"incoming request","method":"GET","path":"/"}
{"time":"2026-08-10T12:00:10Z","level":"ERROR","msg":"manual panic triggered","client_ip":"127.0.0.1"}
```

---

## 4. Konfigurasi Level Log via Environment Variable

Untuk fleksibilitas, server modern biasanya memungkinkan level log diubah tanpa restart (cukup via `kubectl edit configmap`):

```go
var logLevel slog.Level
switch os.Getenv("LOG_LEVEL") {
    case "DEBUG": logLevel = slog.LevelDebug
    case "WARN":  logLevel = slog.LevelWarn
    case "ERROR": logLevel = slog.LevelError
    default:      logLevel = slog.LevelInfo
}
```

---

## Ringkasan Modul 03

- Gunakan **JSON Structured Logging** (bukan `fmt.Println` atau log teks biasa).
- Tulis langsung ke **`stdout`** container.
- Tambahkan **Level Log** (`INFO`, `WARN`, `ERROR`) untuk kemampuan filter yang presisi menggunakan **LogQL**.
