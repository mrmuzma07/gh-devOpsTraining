# Modul 05: Lab Simulasi Insiden & Helm Rollback

> **Target Pembelajaran:** Mensimulasikan insiden kegagalan deployment akibat kesalahan nilai konfigurasi (*human error* pada `values.yaml`), melacak riwayat revisi, serta memulihkan sistem secara instan menggunakan `helm rollback`.

---

## 1. Skenario Insiden

**Kronologi Kejadian:**
> *"Seorang engineer tidak sengaja melakukan upgrade Helm Release menggunakan file konfigurasi yang salah (`forceUnready: "true"` dan alokasi memori terlalu rendah). Akibatnya seluruh Pod di produksi berstatus Unready dan trafik terputus!"*

Sebagai DevOps Engineer di tempat kerja, Anda diminta melakukan **Emergency Rollback** untuk mengembalikan sistem ke kondisi sehat sebelumnya.

---

## 2. Langkah 1: Mensimulasikan Insiden (Bad Deployment Upgrade)

1. **Buat file konfigurasi yang rusak `minggu-03/values-broken.yaml`:**

   ```yaml
   config:
     appEnv: "production-broken"
     forceUnready: "true" # Pemicu kegagalan Readiness Probe

   resources:
     limits:
       memory: "4Mi" # Alokasi RAM terlalu kecil (Rawan OOMKilled)
   ```

2. **Jalankan `helm upgrade` menggunakan file rusak tersebut:**
   ```bash
   helm upgrade go-app-dev ./charts/go-app \
     --namespace mini-prod \
     -f ./values-broken.yaml
   ```

---

## 3. Langkah 2: Observasi Dampak Insiden

1. **Cek Status Pod:**
   ```bash
   kubectl get pods -n mini-prod
   ```
   *Seluruh Pod berada dalam kondisi `0/1 READY` atau `CrashLoopBackOff`.*

2. **Cek Riwayat Revisi Helm:**
   ```bash
   helm history go-app-dev -n mini-prod
   ```
   **Output:**
   ```text
   REVISION  UPDATED                  STATUS      CHART         APP VERSION  DESCRIPTION
   1         Mon Aug 10 18:00:00 2026 superceded go-app-0.1.0  1.0.0        Install complete (Dev)
   2         Mon Aug 10 18:05:00 2026 superceded go-app-0.1.0  1.0.0        Upgrade values-prod
   3         Mon Aug 10 18:10:00 2026 deployed   go-app-0.1.0  1.0.0        Upgrade values-broken
   ```
   *Perhatikan bahwa **Revision 3** aktif, tetapi Pod di cluster mengalami kegagalan.*

---

## 4. Langkah 3: Eksekusi Emergency Rollback

Dari tabel `helm history`, kita tahu bahwa **Revision 2** adalah kondisi produksi yang sehat dan stabil.

Jalankan perintah **Rollback** ke Revision 2:

```bash
helm rollback go-app-dev 2 -n mini-prod
```

**Output Terminal:**
```text
Rollback release go-app-dev to revision 2 was a success! Happy Helming!
```

---

## 5. Langkah 4: Verifikasi Pemulihan Sistem

1. **Cek Riwayat Revisi Terbaru:**
   ```bash
   helm history go-app-dev -n mini-prod
   ```
   **Output:**
   ```text
   REVISION  UPDATED                  STATUS      CHART         APP VERSION  DESCRIPTION
   1         Mon Aug 10 18:00:00 2026 superceded go-app-0.1.0  1.0.0        Install complete
   2         Mon Aug 10 18:05:00 2026 superceded go-app-0.1.0  1.0.0        Upgrade values-prod
   3         Mon Aug 10 18:10:00 2026 superceded go-app-0.1.0  1.0.0        Upgrade values-broken
   4         Mon Aug 10 18:12:00 2026 deployed   go-app-0.1.0  1.0.0        Rollback to 2
   ```
   *Helm membuat **Revision 4** yang isinya persis mengkloning konfigurasi Revision 2.*

2. **Cek Status Pod:**
   ```bash
   kubectl get pods -n mini-prod
   ```
   *Dalam hitungan detik, seluruh Pod kembali sehat `1/1 READY`!*

---

## 6. Best Practice: Menggunakan Flag `--atomic`

Untuk mencegah insiden di atas terjadi sama sekali di masa depan, selalu gunakan flag **`--atomic`** dan **`--timeout`** saat melakukan `helm upgrade`.

```bash
helm upgrade go-app-dev ./charts/go-app \
  --namespace mini-prod \
  -f ./values-broken.yaml \
  --atomic \
  --timeout 1m
```

### Cara Kerja `--atomic`:
1. Helm akan mencoba meng-apply perubahan.
2. Helm menunggu hingga Pod dinyatakan `1/1 READY` dalam batas waktu timeout (1 menit).
3. Jika Pod gagal `READY` dalam 1 menit, **Helm secara otomatis membatalkan transaksi dan melakukan rollback sendiri** tanpa perlu intervensi manual dari Anda!
