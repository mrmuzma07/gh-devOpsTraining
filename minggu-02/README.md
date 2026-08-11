# Minggu 2 — Kubernetes Workload

Selamat datang di materi pembelajaran **Minggu 2: Kubernetes Workload**.

Modul minggu ini berfokus pada bagaimana menjalankan aplikasi tingkat produksi (*production-ready*) di atas Kubernetes. Anda akan mempelajari berbagai jenis workload controller, memisahkan konfigurasi sensitif & non-sensitif (`ConfigMap` & `Secret`), memasang media penyimpanan persisten (`PVC`), memasang mekanisme pemantauan kesehatan aplikasi (`Probes`), serta menangani insiden nyata di Kubernetes.

---

## 🗺️ Struktur & Daftar Modul Pembelajaran

Materi Minggu 2 dibagi menjadi 5 modul dokumen, 1 direktori manifest YAML, dan 1 direktori source code Go:

| No | Dokumen Modul | Deskripsi Topik | Jenis |
| :---: | :--- | :--- | :---: |
| 01 | [Modul 01: Jenis Workload Kubernetes](./01-jenis-workload-kubernetes.md) | ReplicaSet, Deployment, StatefulSet, DaemonSet, Job, CronJob | 📖 Teori |
| 02 | [Modul 02: Konfigurasi & Storage](./02-konfigurasi-dan-storage.md) | ConfigMap, Secret (Base64), PVC, & StorageClass | 📖 Teori |
| 03 | [Modul 03: Health Checks (Probes)](./03-health-checks-probes.md) | Startup Probe, Liveness Probe, & Readiness Probe | 📖 Teori |
| 04 | [Modul 04: Lab Deploy Go App](./04-lab-deploy-go-app.md) | Build Go App image + deploy lengkap (ConfigMap+Secret+PVC+Probes) | 🧪 Lab |
| 05 | [Modul 05: Lab Incident Troubleshooting](./05-lab-incident-troubleshooting.md) | Simulasi Readiness failure, analisa via `kubectl describe/logs` | 🧪 Lab |
| 📂 | [Source Code Go App](./app/) | Kode `main.go` & `Dockerfile` multi-stage build | 💻 Code |
| 📂 | [Manifests YAML](./manifests/) | Manifest `01-configmap`, `02-secret`, `03-pvc`, `04-deployment`, `05-service` | 💻 Code |

---

## 🎯 Target & Checklist Capaian Pembelajaran

- [ ] Memahami kapan harus memilih `Deployment`, `StatefulSet`, `DaemonSet`, `Job`, atau `CronJob`.
- [ ] Mampu memisahkan konfigurasi aplikasi non-sensitif menggunakan **ConfigMap**.
- [ ] Mampu mengelola kredensial & rahasia menggunakan **Secret** (Base64).
- [ ] Memahami cara kerja penyimpanan data persisten dengan **PersistentVolumeClaim (PVC)** dan **StorageClass** `local-path`.
- [ ] Mampu mengonfigurasi **Startup**, **Liveness**, dan **Readiness Probe** pada file YAML Deployment.
- [ ] Berhasil mem-build OCI Image Go App dengan Podman dan mem-deploy-nya ke cluster `k3s`.
- [ ] Mampu melakukan troubleshooting insiden `Readiness Probe Failed` menggunakan `kubectl describe` dan `kubectl logs`.

---

## 🚀 Ringkasan Arsitektur Lab Minggu 2

```
   [ConfigMap: APP_ENV] ──────────┐
   [Secret: DB_USER/PASS] ────────┼──>  [Deployment: Go App]
   [PVC: local-path 1Gi] ─────────┤     ├── Startup Probe (/healthz)
                                  │     ├── Liveness Probe (/healthz)
                                  │     └── Readiness Probe (/ready)
                                  │                │
                                  ▼                ▼
                         [Service: ClusterIP] ──> Endpoints
```

---

## ➡️ Perjalanan Selanjutnya (Persiapan Minggu 3)

Setelah menguasai Kubernetes Workload secara manual tanpa Helm, Anda siap masuk ke **Minggu 3: GitOps & Package Management**, di mana kita akan mempelajari:
- Pengemasan manifest dengan **Helm Chart**
- Konsep **GitOps** menggunakan **ArgoCD**
- Automasi sinkronisasi dari Git Repository langsung ke Cluster k3s!
