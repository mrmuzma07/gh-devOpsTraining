# Tahap 0 — Prasyarat SRE dan DevOps

Tahap ini adalah jembatan sebelum masuk ke **Tahap 1: Fondasi Platform**. Materinya ditujukan untuk pemula yang belum terbiasa dengan terminal, Linux, jaringan, atau workflow pengembangan perangkat lunak.

Tidak perlu menguasai semua topik secara mendalam sebelum mulai. Gunakan checklist dan acceptance gate di setiap modul untuk memastikan fondasi yang dibutuhkan sudah cukup kuat.

## Tujuan Pembelajaran

Setelah menyelesaikan Tahap 0, Anda diharapkan mampu:

- Bekerja nyaman dari terminal dan memahami filesystem serta proses Linux.
- Menjelaskan cara kerja jaringan dasar dan mendiagnosis koneksi sederhana.
- Menggunakan Git dengan aman untuk membuat perubahan, branch, commit, dan pull request.
- Memahami format data dan otomasi dasar yang sering muncul di DevOps.
- Menjalankan aplikasi lokal, membaca dokumentasi, dan mengumpulkan bukti saat troubleshooting.

## Urutan Modul

| No | Modul | Fokus | Hasil praktik |
| :---: | :--- | :--- | :--- |
| 01 | [Linux dan Terminal](./01-linux-dan-terminal.md) | Shell, filesystem, permission, proses, service, package manager, dan troubleshooting dasar | Skrip inspeksi sistem dan catatan diagnosis |
| 02 | [Networking Dasar](./02-networking-dasar.md) | IP, DNS, port, TCP/UDP, HTTP, routing, proxy, dan tools diagnosis | Laporan alur request dan hasil uji konektivitas |
| 03 | [Git dan Workflow Kolaborasi](./03-git-dan-workflow.md) | Repository, commit, branch, merge, rebase, remote, pull request, dan recovery | Repository latihan dengan riwayat perubahan yang rapi |
| 04 | [Otomasi dan Tooling Dasar](./04-otomasi-dan-tooling.md) | YAML, JSON, environment variable, shell script, CLI, dan membaca dokumentasi | Skrip konfigurasi aman dan command reference pribadi |
| 05 | [Praktik Troubleshooting](./05-praktik-troubleshooting.md) | Hipotesis, observasi, perubahan terkontrol, rollback, dan dokumentasi insiden | Mini runbook dan postmortem singkat |

## Cara Belajar

1. Kerjakan modul secara berurutan; modul 01-03 adalah fondasi minimum.
2. Gunakan mesin virtual atau environment lab bila mencoba perintah yang mengubah service, permission, atau konfigurasi jaringan.
3. Jangan menyalin credential asli ke terminal history, repository, screenshot, atau file latihan.
4. Simpan hasil praktik di repository latihan, bukan di branch utama repository materi ini.
5. Jika sebuah konsep sudah dikuasai, gunakan acceptance gate sebagai validasi dan lanjutkan ke modul berikutnya.

## Acceptance Gate Tahap 0

Sebelum masuk ke Tahap 1, pastikan Anda dapat:

- [ ] Menjelaskan lokasi file, permission, proses, dan exit code dari perintah Linux.
- [ ] Menggunakan `ssh`, `curl`, `dig` atau `nslookup`, `ss` atau `lsof`, dan `traceroute` atau `tracepath` untuk diagnosis dasar.
- [ ] Menjelaskan hubungan DNS, IP, port, TCP, HTTP, dan TLS pada sebuah request.
- [ ] Membuat branch, commit atomik, merge atau rebase, menangani conflict, dan membatalkan perubahan dengan aman.
- [ ] Membaca serta mengubah YAML/JSON tanpa merusak struktur atau membocorkan secret.
- [ ] Menulis langkah reproduksi, hipotesis, bukti, mitigasi, dan verifikasi untuk masalah sederhana.
- [ ] Menjalankan aplikasi atau service lokal, lalu membersihkan resource yang dibuat.

Jika semua checklist terpenuhi, lanjutkan ke [Tahap 1 di root roadmap](../README.md#tahap-1-fondasi-platform).
