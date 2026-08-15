# Modul 05 — Praktik Troubleshooting

## Tujuan

Membangun kebiasaan operasi yang terukur sebelum menghadapi incident simulation di Tahap 2. Troubleshooting bukan menebak command, melainkan mengurangi ketidakpastian dengan bukti.

## Metode 6 Langkah

1. **Definisikan gejala** — siapa terdampak, sejak kapan, dan apa yang berubah.
2. **Tentukan scope** — satu proses, satu host, satu service, atau semua pengguna.
3. **Buat hipotesis** — urutkan berdasarkan kemungkinan dan dampak; jangan mengubah banyak hal sekaligus.
4. **Kumpulkan bukti** — status, log, metric, konfigurasi, dependency, dan waktu kejadian.
5. **Lakukan mitigasi terkontrol** — pilih perubahan yang dapat dibalik dan catat command yang dijalankan.
6. **Verifikasi dan dokumentasikan** — pastikan gejala pulih, pantau regresi, lalu tulis follow-up.

## Template Mini Runbook

```text
Judul:
Dampak:
Gejala dan waktu mulai:
Scope:
Hipotesis:
Perintah/observasi:
Mitigasi dan alasan:
Hasil verifikasi:
Rollback:
Follow-up:
```

## Latihan

1. Jalankan service HTTP lokal atau container sederhana.
2. Ubah satu kondisi pada environment lab, misalnya port atau environment variable, lalu amati gejalanya.
3. Diagnosis menggunakan log, process list, `ss`, `curl`, dan konfigurasi; jangan langsung menghapus resource.
4. Pulihkan kondisi, ulangi request, dan tulis postmortem satu halaman: timeline, root cause, dampak, serta pencegahan.

## Acceptance Gate

- [ ] Dapat membedakan fakta, asumsi, dan hipotesis.
- [ ] Dapat mengumpulkan bukti sebelum melakukan perubahan.
- [ ] Dapat menjelaskan blast radius dan langkah rollback.
- [ ] Dapat memverifikasi pemulihan dengan sinyal yang relevan.
- [ ] Dapat menulis runbook yang bisa diikuti orang lain.

## Etika dan keselamatan

Gunakan hanya environment lab. Hindari eksperimen destruktif pada sistem bersama, jangan menguji credential asli, dan hentikan latihan bila dampaknya tidak lagi terkontrol.
