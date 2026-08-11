# Modul 01: Konsep Helm — Kubernetes Package Manager

> **Target Pembelajaran:** Memahami alasan mengapa Kubernetes membutuhkan Package Manager (Helm), perbedaan antara **Chart**, **Values**, dan **Release**, serta keunggulan arsitektur Helm 3.

---

## 1. Mengapa Butuh Helm?

Pada Minggu 1 dan 2, kita mengelola file YAML manifest secara manual (`kubectl apply -f manifest.yaml`).

Namun dalam lingkungan skala besar (*enterprise*), muncul tantangan berikut:
1. **Duplikasi YAML (Copy-Paste Nightmare):** Bagaimana jika kita ingin meng-deploy aplikasi yang sama ke environment `Dev`, `Staging`, dan `Production` dengan beda CPU, RAM, dan Replikasi? Apakah harus membuat 3 folder berisi puluhan file YAML yang mirip?
2. **Sulit Tracking Versi Deployment:** Jika versi aplikasi di-upgrade, bagaimana cara tahu versi berapa yang sedang berjalan di cluster secara resmi?
3. **Rollback Rumit:** Jika deployment baru crash, bagaimana cara membalikkan seluruh state aplikasi (Deployment, Service, ConfigMap) ke versi sebelumnya dalam 1 perintah?

```
Tanpa Helm (Manual YAML statis)           Dengan Helm (Package Manager Dinamis)
┌─────────────────────────────────┐       ┌─────────────────────────────────┐
│ deployment-dev.yaml             │       │                                 │
│ deployment-staging.yaml         │  ==>  │  Chart: templates/ + values.yaml│
│ deployment-prod.yaml            │       │  (1 Template untuk semua Env!)   │
│ (Rentan human error & copy-paste│       │                                 │
└─────────────────────────────────┘       └─────────────────────────────────┘
```

**Solusi Helm:**
Helm bertindak sebagai **Package Manager** (seperti `apt` di Ubuntu, `brew` di macOS, atau `npm` di Node.js) khusus untuk Kubernetes.

---

## 2. Tiga Konsep Utama Helm

Helm dibangun di atas 3 komponen fundamental:

```mermaid
graph LR
    Chart[1. Chart\nTemplate Blueprint] + Values[2. Values.yaml\nVariabel Parameter] -->|helm install| Release[3. Release\nInstansi Aktif di Cluster]
```

1. **Chart (Paket Blueprint):**
   - Berisi kumpulan berkas template YAML yang menentukan struktur resource Kubernetes (Deployment, Service, Ingress, dll).

2. **Values (Konfigurasi Variabel):**
   - Berkas berisi parameter nilai (`values.yaml`) yang akan disuntikkan ke dalam berkas template Chart.
   - Anda bisa membuat file `values-dev.yaml` dan `values-prod.yaml` dari 1 Chart yang sama!

3. **Release (Instansi Running):**
   - Setiap kali sebuah Chart di-install ke cluster dengan kombinasi Values tertentu, Helm membuat sebuah **Release**.
   - Contoh: Dari Chart `go-app`, Anda bisa membuat 2 Release terpisah: `go-app-dev` di namespace `dev` dan `go-app-prod` di namespace `prod`.

---

## 3. Arsitektur Helm 3 (Tillerless)

Penting untuk diketahui bahwa Helm mengalami evolusi besar dari versi 2 ke versi 3:

- **Helm v2 (Legacy):** Menggunakan komponen server bernama `Tiller` yang berjalan sebagai `root` di dalam cluster. Sangat rentan dari sisi keamanan (*security risk*).
- **Helm v3 (Modern):** **Tanpa Tiller (*Tillerless*)**. Helm CLI berkomunikasi langsung dengan Kubernetes API Server menggunakan kredensial `kubeconfig` user lokal.

```
Helm v3 Architecture (Tillerless & Secure)

┌──────────────────────┐   HTTPS / gRPC    ┌──────────────────────┐
│ Helm CLI             │ ────────────────> │ Kubernetes API       │
│ (di Laptop / CI/CD)  │  (via Kubeconfig) │ Server (k3s)         │
└──────────────────────┘                   └──────────────────────┘
```

---

## 4. Perintah Dasar CLI Helm

Berikut adalah perintah CLI Helm yang sering digunakan:

| Perintah | Fungsi |
| :--- | :--- |
| `helm create <nama-chart>` | Membuat struktur folder Helm Chart baru dari nol. |
| `helm repo add <nama> <url>` | Menambahkan repository Helm publik (seperti Artifact Hub). |
| `helm install <release-name> <chart>` | Meng-deploy Chart ke cluster Kubernetes. |
| `helm list` | Menampilkan daftar Release aktif di cluster. |
| `helm upgrade <release-name> <chart>` | Memperbarui Release ke versi Chart/Values baru. |
| `helm rollback <release-name> <revision>` | Membalikkan status Release ke revisi sebelumnya. |
| `helm uninstall <release-name>` | Menghapus Release beserta seluruh resourcenya dari cluster. |

---

## Ringkasan Modul 01

- **Helm** adalah Package Manager Kubernetes yang mengeliminasi duplikasi berkas YAML statis.
- **Chart** = Blueprint template, **Values** = Nilai konfigurasi, **Release** = Hasil instansi aktif di cluster.
- **Helm 3** sepenuhnya aman (*Tillerless*) dan terintegrasi langsung dengan RBAC Kubernetes.
