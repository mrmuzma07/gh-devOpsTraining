# Minggu 3 — Helm (Kubernetes Package Manager)

Selamat datang di materi pembelajaran **Minggu 3: Helm**.

Di minggu ini, kita melangkah dari pengelolaan manifest statis manual ke pengemasan aplikasi dinamis menggunakan **Helm Chart**. Anda akan mengonversi manifest Go App dari Minggu 2 menjadi Helm Chart reusable, memisahkan nilai variabel untuk environment `Dev` dan `Prod`, serta menguasai manajemen revisi release (`helm install`, `upgrade`, dan `rollback`).

---

## 🗺️ Struktur & Daftar Modul Pembelajaran

Materi Minggu 3 dibagi menjadi 5 modul dokumen dan 1 direktori Helm Chart:

| No | Dokumen Modul | Deskripsi Topik | Jenis |
| :---: | :--- | :--- | :---: |
| 01 | [Modul 01: Konsep Helm Package Manager](./01-konsep-helm-package-manager.md) | Helm v3, Chart vs Values vs Release, Tillerless architecture | 📖 Teori |
| 02 | [Modul 02: Anatomi Chart & Templating](./02-struktur-chart-dan-templating.md) | `Chart.yaml`, Go template syntax, `.Values`, `_helpers.tpl` | 📖 Teori |
| 03 | [Modul 03: Manajemen Release Helm](./03-manajemen-release-helm.md) | Lifecycle: `install`, `upgrade`, `history`, `rollback`, `--atomic` | 📖 Teori |
| 04 | [Modul 04: Lab Konversi Go App ke Helm](./04-lab-konversi-go-app-ke-helm.md) | Hands-on konversi manifest ke `charts/go-app` (Dev vs Prod) | 🧪 Lab |
| 05 | [Modul 05: Lab Incident & Rollback](./05-lab-incident-rollback.md) | Simulasi salah values, lacak `helm history`, Emergency Rollback | 🧪 Lab |
| 📂 | [Helm Chart `charts/go-app/`](./charts/go-app/) | Chart template lengkap + `values-dev.yaml` & `values-prod.yaml` | 💻 Code |

---

## 🎯 Target & Checklist Capaian Pembelajaran

- [ ] Memahami 3 konsep utama Helm: **Chart**, **Values**, dan **Release**.
- [ ] Memahami perbedaan arsitektur Helm 2 vs Helm 3 (*Tillerless*).
- [ ] Mampu menulis sintaks Go Templating Helm (`.Values`, `.Release`, `nindent`, `quote`).
- [ ] Mampu membuat dan menggunakan **Named Templates** pada file `_helpers.tpl`.
- [ ] Berhasil mengonversi berkas YAML statis Go App menjadi Helm Chart di `charts/go-app/`.
- [ ] Mampu meng-deploy Chart menggunakan `values-dev.yaml` dan meng-upgradenya menggunakan `values-prod.yaml`.
- [ ] Mampu memulihkan deployment yang rusak menggunakan `helm rollback` dan memahami kegunaan flag `--atomic`.

---

## 🚀 Ringkasan Alur Deployment Helm

```
 [charts/go-app/] + [values-dev.yaml]  ──>  helm install  ──>  Release: go-app-dev (Rev 1)
                                                                       │
 [charts/go-app/] + [values-prod.yaml] ──>  helm upgrade  ──>  Release: go-app-dev (Rev 2)
                                                                       │
 [values-broken.yaml]                  ──>  helm upgrade  ──>  Release: FAILED (Rev 3)
                                                                       │
                                            helm rollback ──>  Release: OK (Rev 4 -> Rev 2)
```

---

## ➡️ Perjalanan Selanjutnya (Persiapan Minggu 4)

Setelah menguasai paket Helm secara lokal di laptop, Anda siap melangkah ke **Minggu 4: GitLab + GitOps (ArgoCD)**, di mana kita akan mempelajari:
- Konsep **GitOps** (Git sebagai Single Source of Truth).
- Menghapuskan penggunaan manual `kubectl apply` / `helm install`.
- Memasang **ArgoCD** di cluster k3s untuk melakukan sinkronisasi otomatis dari Git Repository!
