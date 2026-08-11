# Minggu 4 — GitLab + GitOps (ArgoCD)

Selamat datang di materi pembelajaran **Minggu 4: GitLab + GitOps**.

Di minggu ini, kita meninggalkan cara manual `kubectl apply` dan beralih penuh ke paradigma **GitOps**. Anda akan mengintegrasikan pipeline CI/CD GitLab dengan **ArgoCD** di cluster k3s, memahami pemisahan peran CI dan CD, serta mempraktikkan fitur otomatisasi *Auto-Sync*, *Auto-Prune*, dan *Self-Healing*.

---

## 🗺️ Struktur & Daftar Modul Pembelajaran

Materi Minggu 4 dibagi menjadi 5 modul dokumen dan berkas konfigurasi pendukung:

| No | Dokumen Modul | Deskripsi Topik | Jenis |
| :---: | :--- | :--- | :---: |
| 01 | [Modul 01: Konsep GitOps & Branching Strategy](./01-konsep-gitops-dan-branching-strategy.md) | 4 Prinsip GitOps, Push vs Pull, & Merge Request Workflow | 📖 Teori |
| 02 | [Modul 02: CI/CD Pipeline & GitLab Integration](./02-cicd-pipeline-dan-gitlab-runner.md) | Pemisahan CI (GitLab) vs CD (ArgoCD) & `.gitlab-ci.yml` | 📖 Teori |
| 03 | [Modul 03: Arsitektur ArgoCD & Sync Policy](./03-arsitektur-argocd-dan-declarative-sync.md) | ArgoCD Architecture, CRD `Application`, Auto-Sync, & Self-Healing | 📖 Teori |
| 04 | [Modul 04: Lab Setup ArgoCD & GitOps](./04-lab-setup-gitops-repository-dan-argocd.md) | Hands-on install ArgoCD di k3s & trigger Auto-Sync via Git Commit | 🧪 Lab |
| 05 | [Modul 05: Lab Incident Self-Healing](./05-lab-incident-drift-detection-self-healing.md) | Simulasi hapus Deployment manual & verifikasi ArgoCD Self-Healing | 🧪 Lab |
| 📂 | [ArgoCD Manifest `argocd/`](./argocd/) | Berkas CRD `application-go-app.yaml` | 💻 Code |
| 📂 | [GitLab CI `apps/`](./apps/) | Contoh berkas pipeline `.gitlab-ci.yml` | 💻 Code |
| 📂 | [Cluster Config `clusters/`](./clusters/) | Override values GitOps untuk environment `dev/` dan `prod/` | 💻 Code |

---

## 🎯 Target & Checklist Capaian Pembelajaran

- [ ] Memahami 4 prinsip utama GitOps (Declarative, Versioned, Pulled Automatically, Continuously Reconciled).
- [ ] Memahami perbedaan keunggulan model **Pull-based (ArgoCD)** dibanding Push-based (CI Runner biasa).
- [ ] Memahami alasan pemisahan repository (*App Repo vs Config Repo*).
- [ ] Mampu membaca dan mengonfigurasi CRD ArgoCD `Application`.
- [ ] Berhasil memasang ArgoCD di cluster k3s lokal dan mengakses Dashboard Web UI.
- [ ] Membuktikan fitur **Auto-Sync** saat file `values.yaml` di-push ke Git.
- [ ] Membuktikan fitur **Self-Healing** saat resource Kubernetes dihapus secara manual dari CLI.

---

## 🚀 Ringkasan Alur GitOps

```
 Developer  ──>  git push  ──>  GitLab CI (Build & Test Image)
                                       │
                                       ▼
                              GitOps Config Repo (Git Commit Image Tag)
                                       │
                                       ▼
                              ArgoCD (Polls Git Repo)
                                       │
                                       ▼ (Auto Sync & Self-Healing)
                              k3s Cluster (Namespace: mini-prod)
```

---

## ➡️ Perjalanan Selanjutnya (Persiapan Minggu 5)

Setelah aplikasi ter-deploy secara otomatis melalui GitOps, Anda siap masuk ke **Minggu 5: Metrics & Observability**, di mana kita akan mempelajari:
- Memasang **Grafana Stack** (Grafana, Alloy, Mimir).
- Pengumpulan metrik dalam format Prometheus.
- Menampilkan grafik monitoring CPU, Memory, dan HTTP Request Rate secara real-time!
