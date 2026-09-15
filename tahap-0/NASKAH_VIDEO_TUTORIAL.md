# 🎬 NASKAH VIDEO TUTORIAL: TAHAP 0 — FONDASI SRE & DEVOPS

> **Format:** Video Tutorial / Crash Course (Seri Pembelajaran DevOps)  
> **Gaya Penyampaian:** Santai, Mengalir, Edukatif, Mudah Dipahami Pemula (No-BS, Direct-to-the-point)  
> **Target Audiens:** Pemula, Junior Engineer, Sysadmin, atau Developer yang ingin masuk ke dunia DevOps/SRE.

---

## 📌 METADATA VIDEO

- **Judul Video (Rekomendasi):**  
  *Panduan Lengkap DevOps untuk Pemula: Dari Zero ke Fondasi SRE (Tahap 0)*
- **Durasi Estimasi:** 25 - 35 Menit
- **Elemen Format:**
  - `[VISUAL]`: Petunjuk tampilan layar, B-Roll, animasi, atau slide/terminal yang ditampilkan.
  - `[AUDIO / VOICEOVER]`: Dialog/kata-kata yang diucapkan pembicara.
  - `[TEXT ON SCREEN]`: Teks/grafik penjelas di layar video.

---

## 📍 BREAKDOWN NASKAH VIDEO

---

### ⏱️ SEGMEN 1: INTRO & MINDSET TAHAP 0 (00:00 - 03:00)

**[VISUAL]**  
- Pembicara di depan kamera (Talking Head) dengan latar belakang setup meja kerja ramah tech.  
- Teks Judul Muncul dengan animasi menarik.

**[TEXT ON SCREEN]**  
*Tahap 0: Prasyarat SRE & DevOps — Fondasi Sebelum Belajar Kubernetes & Cloud!*

**[AUDIO / VOICEOVER]**  
"Halo semuanya! Selamat datang di seri tutorial DevOps dan SRE.

Kalau kamu baru mau terjun ke dunia DevOps, mungkin kamu sering dengar istilah keren seperti *Kubernetes, Docker, CI/CD, Terraform*, atau *Grafana*. Tapi pas mau belajar, malah bingung harus mulai dari mana...

Nah, kesalahan terbanyak pemula adalah **langsung lompat ke tool canggih tanpa punya fondasi dasar yang kuat.** Hasilnya? Pas aplikasi error di server, bingung cari tahu penyebabnya di mana.

Makanya, di video ini kita bakal bahas **Tahap 0: Prasyarat Utama SRE dan DevOps**. Ini adalah 5 fondasi wajib yang bakal bikin perjalanan belajarmu di dunia cloud-native jauh lebih mulus.

Kita bakal bedah 5 topik utama:
1. Linux & Terminal CLI
2. Networking Dasar & Diagnostic Tools
3. Git & Workflow Kolaborasi
4. Otomasi, Config (YAML/JSON) & Shell Scripting
5. Metodologi Troubleshooting SRE

Siapkan kopi atau catatanmu, mari kita mulai dari topik pertama!"

---

### ⏱️ SEGMEN 2: MODUL 01 — LINUX & TERMINAL CLI (03:00 - 08:30)

**[VISUAL]**  
- Transisi layar ke rekaman layar (Screen Recording) Terminal Linux.  
- Diagram hirarki Linux Filesystem (`/`).

**[TEXT ON SCREEN]**  
*Modul 01: Linux & CLI Terminal — Lapangan Bermain Harian DevOps*

**[AUDIO / VOICEOVER]**  
"Hampir 99% server produksi, runner CI/CD, dan node Kubernetes itu berjalan di atas sistem operasi **Linux**. Di server produksi, nggak ada yang namanya tampilan antarmuka grafis (GUI) kaya Windows atau macOS. Semuanya serba hitam-putih di Terminal.

Bahkan perlu diingat: **Container sekelas Docker itu sebenarnya cuma proses Linux biasa** yang diisolasi pakai fitur Kernel bernama *Namespaces* dan *Cgroups*.

Jadi, apa aja yang wajib kamu kuasai di Linux?"

**[VISUAL]**  
- Membuka terminal, mengetik perintah navigasi dasar (`pwd`, `ls -la`, `cd`).  
- Diagram folder Linux: `/`, `/etc`, `/var/log`, `/usr/bin`.

**[AUDIO / VOICEOVER]**  
"Pertama, **Struktur Filesystem Linux**. Bedanya sama Windows yang pake drive `C:` atau `D:`, Linux itu cuma punya satu akar yaitu root atau slash (`/`).
- Folder `/etc` adalah tempat nangkring semua file konfigurasi.
- Folder `/var/log` tempat tersimpannya catatan atau log sistem dan aplikasi.
- Folder `/usr/bin` tempat perintah atau executable berada.

Coba ingat perintah dasar navigasi ini: `ls -la` buat ngeliat isi folder lengkap beserta file tersembunyi, `pwd` buat ngecek kita lagi ada di folder mana, dan `cd` buat pindah folder."

**[VISUAL]**  
- Praktik di terminal: Menunjukkan `ls -l` dan menjelaskan kolom permission `rwxr-xr--` dan `chmod`/`chown`.

**[AUDIO / VOICEOVER]**  
"Kedua, **File Permission & Ownership**. Sering banget aplikasi error gara-gara *Permission Denied*.  
Di Linux, izin akses dibagi 3: **Read (r)**, **Write (w)**, dan **Execute (x)** untuk User, Group, dan Other.  
Kalau mau ubah izin file, kita pakai `chmod`. Kalau mau ganti pemilik file, kita pakai `chown`."

**[VISUAL]**  
- Mengosongkan layar (`clear`), lalu menjalankan `systemctl status` atau `journalctl -u`.  
- Membuka `top` atau `htop` di terminal.

**[AUDIO / VOICEOVER]**  
"Ketiga, **Proses & Service Management**.  
Untuk mengelola layanan di background (seperti web server Nginx), sistem Linux modern menggunakan `systemd`.  
- Jalankan `systemctl status nama-service` buat ngecek statusnya.
- Mau lihat log aplikasimu? Gunakan `journalctl -u nama-service -f` buat nge-track log secara realtime!
- Mau ngecek CPU atau RAM yang kepakai? Tinggal ketik `top` atau `htop`.

Kuasai ini dulu, dan kamu udah 50% lebih pede buat masuk ke server Linux!"

---

### ⏱️ SEGMEN 3: MODUL 02 — NETWORKING DASAR & DIAGNOSTIC TOOLS (08:30 - 14:00)

**[VISUAL]**  
- Transisi slide/diagram interaktif: Alur Request Jaringan (Client -> DNS -> Router -> Server).  
- Teks langkah 1 sampai 6 berpendar bergantian.

**[TEXT ON SCREEN]**  
*Modul 02: Networking Fundamental & Diagnostic Tools*

**[AUDIO / VOICEOVER]**  
"Sekarang kita masuk ke topik yang paling sering bikin pusing: **Networking**.

Bayangkan aplikasi Node.js kamu mau panggil API di `https://api.example.com/healthz`. Apa yang sebenarnya terjadi di belakang layar?"

**[VISUAL]**  
- Diagram memperlihatkan alur:
  1. DNS Lookup (Mencari IP dari domain)
  2. TCP 3-Way Handshake (SYN -> SYN-ACK -> ACK)
  3. TLS Handshake (Enkripsi HTTPS)
  4. HTTP Request (GET /healthz) & Response 200 OK.

**[AUDIO / VOICEOVER]**  
"Alurnya begini:
1. Laptop kamu bakal nanya ke **DNS**: *'IP-nya api.example.com berapa ya?'* DNS jawab: *'104.21.12.34'*.
2. Laptop kamu bikin koneksi awal via **TCP 3-Way Handshake**.
3. Kalau pakai HTTPS, lanjut ke **TLS Handshake** buat enkripsi keamanan.
4. Baru deh request HTTP dikirim dan server balas pakai respon `200 OK`.

Kenapa alur ini penting? Supaya pas ada error, kamu tahu di tahap mana koneksi itu putus!
- Kalau errornya `Could not resolve host`, berarti DNS-nya bermasalah.
- Kalau `Connection Refused` atau `Timed Out`, berarti Port atau Firewall-nya bermasalah.
- Kalau dapet `500 Internal Server Error`, berarti aplikasinya yang crash."

**[VISUAL]**  
- Demo praktis di terminal dengan tools: `ping`, `dig`, `curl -v`, `ss -tulpn`.

**[AUDIO / VOICEOVER]**  
"Sebagai DevOps, kamu harus jago pakai **5 Tools Diagnosis CLI wajib** ini:
1. `ping`: Buat tes konektivitas fisik dasar (apakah server hidup?).
2. `dig` atau `nslookup`: Buat tes resolusi DNS.
3. `curl -v`: Tool andalan buat kirim request HTTP dan ngeliat detail header serta status code. Pake flag `-v` buat ngeliat proses handshakenya!
4. `ss -tulpn` atau `netstat`: Buat ngeliat port berapa aja yang lagi mendengarkan (*listening*) di server.
5. `traceroute` / `mtr`: Buat ngelacak jalur paket data dari laptopmu ke server tujuan.

Dengan 5 alat ini, kamu nggak bakal menebak-nebak lagi pas ada isu jaringan!"

---

### ⏱️ SEGMEN 4: MODUL 03 — GIT & WORKFLOW KOLABORASI (14:00 - 19:30)

**[VISUAL]**  
- Animasi sederhana diagram GitOps: Developer -> Git Push -> GitHub -> ArgoCD / Server Deployment.

**[TEXT ON SCREEN]**  
*Modul 03: Git & Workflow DevOps — Single Source of Truth*

**[AUDIO / VOICEOVER]**  
"Di dunia modern, Git itu bukan cuma tempat nyimpen kode aplikasi buatan developer. Di DevOps, Git adalah **Single Source of Truth** atau satu-satunya sumber kebenaran.

Semua konfigurasi infrastruktur, manifest Kubernetes, dan pipeline CI/CD disimpan di Git. Konsep ini dikenal sebagai **GitOps**."

**[VISUAL]**  
- Diagram 4 Area Git: Working Tree -> Staging Area -> Local Repository -> Remote Repository (GitHub).  
- Demo terminal melakukan `git add`, `git commit`, `git push`, `git checkout -b feature-xxx`.

**[AUDIO / VOICEOVER]**  
"Pahami **Model Mental 4 Area Git** berikut:
1. **Working Tree**: File yang lagi kamu edit di editor.
2. **Staging Area**: Draf file yang udah kamu pilih pakai `git add`.
3. **Local Repo**: Riwayat commit yang tersimpan di komputer kamu (folder `.git`).
4. **Remote Repo**: Repository utama yang ada di GitHub atau GitLab.

Praktik terbaiknya:
- **Jangan pernah push langsung ke branch `main`!**
- Selalu buat branch fitur baru: `git checkout -b feature/tambah-login`.
- Setelah selesai, kirim perubahan lewat **Pull Request (PR)** supaya bisa di-review oleh rekan tim.
- Jika terjadi *Git Conflict*, jangan panik. Buka filenya, cari penanda conflict, pilih kode mana yang benar, lalu commit ulang."

---

### ⏱️ SEGMEN 5: MODUL 04 — OTOMASI, YAML/JSON & SHELL SCRIPTING (19:30 - 24:30)

**[VISUAL]**  
- Membuka VS Code / Editor yang memperlihatkan file YAML dan JSON bersisian.  
- Menonjolkan perbedaan indentasi spasi vs kurung `{}`.

**[TEXT ON SCREEN]**  
*Modul 04: Otomasi, Format Data (YAML/JSON), & Shell Scripting*

**[AUDIO / VOICEOVER]**  
"Dalam pekerjaan sehari-hari, kamu bakal sering banget berhubungan dengan dua format data ini: **YAML** dan **JSON**.

- **JSON** biasanya dipakai untuk API Payload atau file konfigurasi aplikasi.
- **YAML** adalah standar utama di dunia DevOps (Docker Compose, Kubernetes, Ansible, GitHub Actions).

⚠️ **Catatan Penting buat Pemula:** YAML itu **sangat sensitif terhadap indentasi SPASI**. Jangan pernah mencampur Spasi dan Tab di YAML, karena salah 1 spasi aja bisa bikin deployment kamu gagal total!"

**[VISUAL]**  
- Menunjukkan contoh Environment Variables di terminal: `export DB_HOST=localhost` dan penggunaan `.env`.  
- Membuka skrip bash sederhana dengan header `#!/bin/bash` dan `set -euo pipefail`.

**[AUDIO / VOICEOVER]**  
"Selain format data, kamu juga perlu paham **Environment Variables** (seperti `PORT`, `DB_URL`) untuk memisahkan konfigurasi dari kode program demi keamanan.

Dan untuk otomasi tugas repetitif, kita pakai **Shell Scripting**.  
Tapi ingat, nulis skrip bash di DevOps itu ada aturannya!  
Selalu tambahkan mantra ajaib ini di awal skrip kamu:
```bash
set -euo pipefail
```
Artinya: jika ada perintah yang error atau variabel yang kosong, skrip akan langsung berhenti dan tidak mengeksekusi perintah berbahaya selanjutnya."

---

### ⏱️ SEGMEN 6: MODUL 05 — METODE TROUBLESHOOTING SRE (24:30 - 29:00)

**[VISUAL]**  
- Pembicara kembali di depan kamera (Talking Head).  
- Teks kutipan muncul di layar: *"Don't Guess, Measure!" (Jangan Menebak, Ukur!)*.

**[TEXT ON SCREEN]**  
*Modul 05: Filosofi Troubleshooting SRE & Incident Response*

**[AUDIO / VOICEOVER]**  
"Topik terakhir di Tahap 0 adalah **Metode Troubleshooting**. Ini yang membedakan engineer senior dan junior.

Prinsip nomor satu SRE saat terjadi insiden: **Berdasarkan Bukti, Bukan Tebakan!**  
Jangan pernah bilang *'Mungkin servernya hang deh, coba restart aja'*. Selalu kumpulkan bukti log dan metrik terlebih dahulu."

**[VISUAL]**  
- Diagram Alur 6 Langkah Troubleshooting SRE:
  1. Detect (Deteksi) -> 2. Triage (Penilaian Dampak) -> 3. Isolate (Isolasi Akar Masalah) -> 4. Mitigate (Pertolongan Pertama) -> 5. Resolve (Perbaikan Permanen) -> 6. Postmortem (Pembelajaran).

**[AUDIO / VOICEOVER]**  
"Gunakan **Metodologi 6 Langkah SRE**:
1. **Detect**: Mengetahui ada masalah dari alert atau laporan user.
2. **Triage**: Seberapa parah dampaknya? Sistem mana yang kena?
3. **Isolate**: Cari akar masalah dengan memeriksa log (`journalctl`), koneksi (`curl`), dan resource (`top`).
4. **Mitigate**: Lakukan pertolongan pertama (misal: rollback ke versi sebelumnya agar user bisa transaksi lagi).
5. **Resolve**: Perbaiki bug utamanya secara permanen.
6. **Postmortem**: Tulis dokumen pembelajaran (tanpa menyalahkan orang / *blameless postmortem*) agar masalah yang sama tidak terulang di masa depan."

---

### ⏱️ SEGMEN 7: OUTRO & NEXT STEPS (29:00 - 31:00)

**[VISUAL]**  
- Pembicara di depan kamera.  
- Teks rangkuman Tahap 0 di layar dan pratinjau materi Tahap 1.

**[TEXT ON SCREEN]**  
*Langkah Selanjutnya: Tahap 1 — Fondasi Platform (Docker, CI/CD, Cloud)*

**[AUDIO / VOICEOVER]**  
"Luar biasa! Kita sudah menuntaskan seluruh fondasi di **Tahap 0: Prasyarat SRE dan DevOps**.

Dengan menguasai:
✅ Navigasi Linux & Pengelolaan Log  
✅ Jaringan Dasar & Tools Diagnostic  
✅ Git & Workflow GitOps  
✅ YAML, Env Vars, & Bash Scripting  
✅ Metodologi Troubleshooting berbasis Bukti  

Kamu sekarang sudah punya paspor yang kuat untuk masuk ke **Tahap 1: Fondasi Platform**, di mana kita bakal mulai mempraktekkan *Containerization dengan Docker, membuat Pipeline CI/CD, dan mengelola Server Cloud!*

Kalau kamu suka dengan video ini, jangan lupa klik tombol **Like**, **Subscribe**, dan bagikan ke teman-temanmu yang sedang belajar DevOps. 

Tulis di kolom komentar kalau ada materi yang ingin kamu bedah lebih dalam. Sampai jumpa di video selanjutnya!"

---

## 💡 TIPS REKAMAN & EDITING VIDEO

1. **Kecepatan Bicara:** Gunakan artikulasi yang jelas dengan tempo sedang. Beri jeda 1-2 detik saat transisi antar modul.
2. **Overlay Kode:** Saat mendemonstrasikan terminal atau file YAML, gunakan font yang besar dan kontras (misal: theme VS Code *One Dark Pro* atau *Tokyo Night*).
3. **Grafik / Highlight:** Gunakan kotak sorot merah/kuning saat menunjukkan perintah penting di layar terminal agar perhatian penonton langsung tertuju ke poin utama.
