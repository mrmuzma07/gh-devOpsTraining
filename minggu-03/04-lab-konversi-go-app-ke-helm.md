# Modul 04: Lab Hands-on — Konversi Go App ke Helm Chart

> **Target Pembelajaran:** Mengonversi manifest statis Minggu 2 menjadi Helm Chart yang dinamis, serta mempraktikkan deployment multi-environment (`values-dev.yaml` & `values-prod.yaml`) menggunakan `helm install` dan `helm upgrade`.

---

## 1. Arsitektur Konversi

Kita mengubah seluruh berkas YAML statis dari Minggu 2 menjadi satu kesatuan Helm Chart di folder `minggu-03/charts/go-app/`:

```text
Minggu 2 (Manifesis Statis)                  Minggu 3 (Dinamis Helm Chart)
├── 01-configmap.yaml         ─────────┐    charts/go-app/
├── 02-secret.yaml            ─────────┼──> ├── Chart.yaml
├── 03-pvc.yaml               ─────────┤    ├── values.yaml (values-dev / values-prod)
├── 04-deployment.yaml        ─────────┤    └── templates/
└── 05-service.yaml           ─────────┘        ├── _helpers.tpl
                                                └── (templates YAML)
```

---

## 2. Langkah Demi Langkah

Navigasikan terminal Anda ke direktori `minggu-03/`:

```bash
cd minggu-03
```

---

### Langkah 1: Linting & Validasi Syntax Chart

Sebelum memasang Chart ke cluster, pastikan tidak ada kesalahan sintaks YAML atau Go Templating:

```bash
# 1. Linting Chart (Cek kesalahan struktur & rekomendasi Helm)
helm lint ./charts/go-app

# 2. Render Template untuk Environment Dev ke stdout
helm template go-app-test ./charts/go-app -f ./charts/go-app/values-dev.yaml
```

---

### Langkah 2: Deploy ke Environment Development (`values-dev.yaml`)

Kita akan menginstal Chart ini sebagai Release bernama **`go-app-dev`** di namespace `mini-prod`:

```bash
helm install go-app-dev ./charts/go-app \
  --namespace mini-prod \
  -f ./charts/go-app/values-dev.yaml
```

**Verifikasi Deployment:**
```bash
# 1. Cek daftar Release Helm
helm list -n mini-prod

# 2. Cek resource Kubernetes yang dibuat
kubectl get pods,svc,pvc,secret,configmap -n mini-prod
```
*Amati bahwa hanya 1 Pod yang dibuat (`replicaCount: 1`), dan secret `DB_USER` terisi `dev_user`.*

---

### Langkah 3: Uji Akses Release Dev

Buka port-forwarding:
```bash
kubectl port-forward svc/go-app-dev 8080:8080 -n mini-prod
```
Di terminal lain, akses endpoint utama:
```bash
curl http://localhost:8080/
```
**Output:**
```json
{
  "message": "Hello from Production-Ready Go App on K8s!",
  "environment": "development",
  "db_user": "dev_user"
}
```

---

### Langkah 4: Promote / Upgrade ke Environment Production (`values-prod.yaml`)

Sekarang bayangkan aplikasi siap dipromosikan ke tingkat Production. Tanpa mengubah template kode YAML sedikitpun, kita cukup meng-upgrade Release menggunakan berkas `values-prod.yaml`:

```bash
helm upgrade go-app-dev ./charts/go-app \
  --namespace mini-prod \
  -f ./charts/go-app/values-prod.yaml
```

---

### Langkah 5: Verifikasi Hasil Upgrade

Periksa perubahan yang terjadi setelah perintah `helm upgrade`:

1. **Cek Jumlah Pod (Replikasi):**
   ```bash
   kubectl get pods -n mini-prod
   ```
   *Jumlah Pod otomatis bertambah dari 1 menjadi 3 Pod!*

2. **Cek Riwayat Revisi Helm:**
   ```bash
   helm history go-app-dev -n mini-prod
   ```
   *Terlihat `REVISION 2` berstatus `deployed`.*

3. **Cek Respon API:**
   ```bash
   curl http://localhost:8080/
   ```
   *Variabel `environment` berubah menjadi `"production"`, dan `db_user` berubah menjadi `"prod_admin"`.*

> **Hebat!** Anda telah berhasil meng-upgrade kapasitas dan konfigurasi aplikasi secara instan hanya dengan memindahkan file values Helm!
