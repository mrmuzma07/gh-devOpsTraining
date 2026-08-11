# Minggu 1 — Fundamental Container & Kubernetes

Selamat datang di materi pembelajaran **Minggu 1: Fundamental Container & Kubernetes**. 

Modul ini dirancang secara terstruktur dan bertahap untuk membantu pemula IT memahami konsep dasar containerization, arsitektur Kubernetes (k3s), perintah CLI `kubectl`, serta membuat deployment pertama tanpa bantuan alat abstraksi seperti Helm.

---

## 🗺️ Struktur & Daftar Modul Pembelajaran

Materi Minggu 1 dibagi menjadi 5 modul dokumen dan 1 direktori manifest YAML:

| No | Dokumen Modul | Deskripsi Topik | Jenis |
| :---: | :--- | :--- | :---: |
| 01 | [Modul 01: Konsep Container & OCI](./01-konsep-container.md) | Container vs VM, OCI, Podman, Image Layer & Lifecycle | 📖 Teori |
| 02 | [Modul 02: Arsitektur Kubernetes & k3s](./02-arsitektur-kubernetes.md) | Control Plane, Worker Node, Pod & Pod Lifecycle | 📖 Teori |
| 03 | [Modul 03: Instalasi & Persiapan Environment](./03-instalasi-persiapan.md) | Panduan instalasi Podman, k3s/k3d, & kubectl | 🛠️ Praktik |
| 04 | [Modul 04: Perintah Dasar CLI kubectl](./04-perintah-dasar-kubectl.md) | Menguasai `get`, `describe`, `logs`, `exec`, & `top` | 🛠️ Praktik |
| 05 | [Modul 05: Lab Hands-on Deploy Nginx](./05-lab-hands-on.md) | Step-by-step deploy Nginx via YAML & Auto-Healing Test | 🧪 Lab |
| 📂 | [Manifests YAML](./manifests/) | File YAML `01-namespace`, `02-deployment`, `03-service`, `04-ingress` | 💻 Code |

---

## 🎯 Target & Checklist Capaian Pembelajaran

Gunakan checklist ini untuk memantau kemajuan belajar Anda minggu ini:

- [ ] Memahami perbedaan fundamental antara Container dan Virtual Machine (VM).
- [ ] Memahami alasan memilih Podman (*Daemonless & Rootless*) dibanding Docker klasik.
- [ ] Memahami fungsi komponen Control Plane (API Server, etcd, Scheduler) dan Worker Node (kubelet).
- [ ] Berhasil menginstal `Podman`, `k3s` (atau `k3d`), dan `kubectl` di laptop lokal.
- [ ] Mampu menjalankan perintah `kubectl get`, `describe`, `logs`, `exec`, dan `top` serta membaca outputnya.
- [ ] Berhasil melakukan deploy Nginx dari file YAML buatan sendiri (`Namespace` → `Deployment` → `Service` → `Ingress`).
- [ ] Membuktikan fitur *Auto-healing* Kubernetes dengan sengaja menghapus Pod Nginx.

---

## 🚀 Ringkasan Alur Lab Minggu 1

```
  [01-namespace.yaml]      -->  Membuat ruang isolasi 'mini-prod'
          │
          ▼
  [02-deployment.yaml]     -->  Menjalankan 2 Pod Nginx
          │
          ▼
  [03-service.yaml]        -->  Menyediakan ClusterIP & Load Balancer internal
          │
          ▼
  [04-ingress.yaml]        -->  Menghubungkan domain http://mini-prod.local dari laptop
```

---

## ➡️ Perjalanan Selanjutnya (Persiapan Minggu 2)

Setelah menyelesaikan Minggu 1 dengan baik, Anda siap melangkah ke **Minggu 2: Kubernetes Workload**, di mana kita akan mempelajari:
- StatefulSet vs DaemonSet vs Job vs CronJob
- Manajemen Konfigurasi & Secret (`ConfigMap` & `Secret`)
- Persistent Storage (`PVC` & `StorageClass`)
- Health Check Aplikasi (`Liveness`, `Readiness`, & `Startup Probe`)
