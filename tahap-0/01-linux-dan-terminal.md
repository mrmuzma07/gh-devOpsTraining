# Modul 01: Fundamental Linux & CLI Terminal

> **Target Pembelajaran:** Memahami lingkungan kerja Linux, struktur filesystem, permission, pengelolaan proses & service (systemd/journalctl), package manager, SSH remote access, serta menulis skrip inspeksi sistem.

---

## 1. Mengapa Linux Sangat Penting bagi SRE & DevOps?

Hampir 99% sistem cloud native, container (Docker/OrbStack), node Kubernetes, runner CI/CD, dan server produksi berjalan di atas **Sistem Operasi Linux**. Sebagai SRE atau DevOps Engineer, Linux adalah "lapangan bermain" harian Anda.

```
┌────────────────────────────────────────────────────────┐
│               Aplikasi (Go, Node.js, Python)           │
├────────────────────────────────────────────────────────┤
│          Container Runtime (Docker / OrbStack / k3s)   │
├────────────────────────────────────────────────────────┤
│  Linux Kernel (Process Management, Namespaces, Cgroups)│
├────────────────────────────────────────────────────────┤
│                 Hardware / VM Cloud Instance           │
└────────────────────────────────────────────────────────┘
```

**Alasan Utama Menguasai Linux:**
- **Container adalah Proses Linux:** Container bukan VM; container adalah proses Linux biasa yang diisolasi menggunakan fitur kernel bernama *Namespaces* dan *Cgroups*.
- **Troubleshooting tanpa GUI:** Di server produksi tidak ada antarmuka grafis (GUI). Seluruh diagnosis dilakukan via Terminal / CLI.
- **Otomasi Scripting:** Tugas repetitif (backup, health check, log rotation) diotomasi menggunakan shell script.

---

## 2. Struktur Filesystem & Navigasi CLI Linux

### A. Hierarki Direktori Linux (Filesystem Hierarchy Standard)

Berbeda dengan Windows yang menggunakan drive letter (`C:\`, `D:\`), Linux menggunakan hirarki pohon tunggal berakar pada `/` (root).

```
/ (Root)
├── bin -> /usr/bin     (Perintah dasar sistem: ls, cp, mv, cat)
├── etc/                (File konfigurasi sistem & aplikasi)
├── home/               (Direktori utama user biasa: /home/username)
├── root/               (Direktori utama superuser root)
├── var/                (Data dinamis & log: /var/log, /var/lib)
├── tmp/                (File sementara / temporary)
├── proc/               (Filesystem virtual status kernel & PID)
└── sys/                (Filesystem virtual perangkat & hardware)
```

### B. Perintah Dasar Navigasi & Manipulasi File

```bash
# 1. Melihat posisi direktori saat ini (Print Working Directory)
pwd

# 2. Menampilkan isi direktori secara detail (l: long format, a: all files, h: human readable size)
ls -lah

# 3. Perpindahan direktori
cd /var/log

# 4. Mencari file berdasarkan nama di dalam direktori /etc
find /etc -maxdepth 2 -name "*.conf"

# 5. Mencari teks kata kunci "error" di dalam file log (case-insensitive)
grep -i "error" /var/log/syslog

# 6. Membaca isi file secara bertahap
less /var/log/syslog

# 7. Membaca 10 baris pertama / terakhir file
head -n 10 /var/log/syslog
tail -n 10 /var/log/syslog
```

**Contoh Output `ls -lah`:**
```text
drwxr-xr-x  5 developer  staff   160B Aug 15 10:00 .
drwxr-xr-x 12 developer  staff   384B Aug 15 09:30 ..
-rw-r--r--  1 developer  staff   1.2K Aug 15 10:05 config.yaml
-rwxr-xr-x  1 developer  staff   450B Aug 15 10:08 deploy.sh
```

### C. Redirection & Pipeline (Penggabungan Perintah)

```bash
# 1. Redirection > (Overwrite file)
echo "LOG_LEVEL=debug" > app.env

# 2. Redirection >> (Append / Tambah baris ke akhir file)
echo "PORT=8080" >> app.env

# 3. Pipeline | (Meneruskan output perintah kiri sebagai input perintah kanan)
cat /var/log/syslog | grep -i "failed" | head -n 5

# 4. Command tee (Menampilkan ke layar sekaligus menyimpan ke file)
df -h | tee disk_report.txt
```

---

## 3. User, Group, & Hak Akses (Permission)

Setiap file dan direktori di Linux memiliki pemilik (*User*), kelompok pemilik (*Group*), serta mode akses (*Permission*).

### A. Anatomi Hak Akses (`rwx`)

```text
- rwx r-x r--   1 owner group  file.txt
│  │   │   └─ Permissions for Others (Read only = 4)
│  │   └───── Permissions for Group  (Read + Execute = 5)
│  └───────── Permissions for Owner  (Read + Write + Execute = 7)
└──────────── File Type (- = File biasa, d = Directory, l = Symlink)
```

| Symbol | Mode | Nilai Octal | Keterangan |
| :---: | :---: | :---: | :--- |
| `r` | Read | 4 | Membaca file / melihat daftar isi direktori |
| `w` | Write | 2 | Mengubah file / membuat & menghapus file di direktori |
| `x` | Execute | 1 | Menjalankan script / masuk ke dalam direktori (`cd`) |

### B. Mengubah Permission & Owner

```bash
# 1. Memberikan akses eksekusi ke script (User + Execute)
chmod u+x script.sh

# 2. Mengubah mode permission ke 755 (rwxr-xr-x)
chmod 755 script.sh

# 3. Mengubah pemilik file menjadi user 'appuser' dan group 'appgroup'
sudo chown appuser:appgroup /var/www/app
```

> 🔒 **Security Best Practice (Least Privilege):**
> Jangan pernah memberikan permission `777` (full access untuk siapapun) di server produksi. Jalankan aplikasi dengan user khusus non-root.

---

## 4. Pengelolaan Proses, Systemd Service, & Log

### A. Memantau Proses Sistem

```bash
# 1. Menampilkan seluruh proses yang sedang berjalan di sistem
ps aux | head -n 10

# 2. Memantau penggunaan CPU & RAM secara interaktif
top

# 3. Mencari Process ID (PID) dari aplikasi Nginx
pgrep -af nginx

# 4. Menghentikan proses secara halus (SIGTERM = 15) atau paksa (SIGKILL = 9)
kill -15 <PID>
kill -9 <PID>

# 5. Memeriksa penggunaan RAM dan Disk
free -h
df -h
```

**Contoh Output `free -h`:**
```text
               total        used        free      shared  buff/cache   available
Mem:           7.7Gi       2.1Gi       3.4Gi       12Mi       2.2Gi       5.3Gi
Swap:          2.0Gi          0B       2.0Gi
```

### B. Pengelolaan Service dengan `systemctl`

Di distro modern (Ubuntu, Debian, CentOS), service dikelola oleh **systemd**.

```bash
# 1. Memeriksa status service
systemctl status nginx

# 2. Memulai, menghentikan, dan restart service
sudo systemctl start nginx
sudo systemctl stop nginx
sudo systemctl restart nginx

# 3. Mengaktifkan service agar otomatis jalan saat boot
sudo systemctl enable nginx
```

### C. Membaca Log dengan `journalctl`

```bash
# 1. Membaca 50 baris log terakhir service Nginx
sudo journalctl -u nginx -n 50 --no-pager

# 2. Live streaming log (seperti tail -f)
sudo journalctl -u nginx -f
```

---

## 5. Package Manager, Environment Variables, & SSH Remote Access

### A. Package Manager

- **Ubuntu / Debian:** `apt update && apt install -y curl git`
- **Alpine Linux (Docker Container):** `apk add --no-cache curl git`
- **RHEL / CentOS:** `dnf install -y curl git`

### B. Environment Variables

Environment Variable digunakan untuk melewatkan konfigurasi ke aplikasi tanpa mengubah kode program.

```bash
# 1. Menampilkan nilai variabel
echo $USER
echo $PATH

# 2. Membuat variabel lingkungan baru untuk sesi terminal saat ini
export APP_ENV=production
export PORT=8080

# 3. Menampilkan seluruh variabel lingkungan aktif
env | grep APP_
```

### C. SSH Remote Access (Akses Remote Server)

```bash
# 1. Membuat pasangan SSH Key (Ed25519 disarankan)
ssh-keygen -t ed25519 -C "developer@example.com"

# 2. Menghubungkan ke server remote via SSH
ssh -i ~/.ssh/id_ed25519 user@192.168.1.100
```

> ⚠️ **Penting:**
> File `id_ed25519` adalah **Private Key**. Jangan pernah membagikan atau mengunggah Private Key ke Git repository!

---

## 6. Lab Hands-on: Membuat Skrip Inspeksi Health Check Sistem

Pada latihan ini, Anda akan membuat shell script sederhana untuk memeriksa kesehatan server lokal.

### Langkah 1: Buat Direktori & File Script

```bash
mkdir -p ~/tahap-0-lab/scripts
cd ~/tahap-0-lab/scripts
cat << 'EOF' > sys-inspect.sh
#!/usr/bin/env bash

echo "=========================================="
echo "      SYSTEM HEALTH INSPECTION REPORT     "
echo "=========================================="
echo "Date/Time : $(date)"
echo "Hostname  : $(hostname)"
echo "User      : $(whoami)"
echo "Uptime    : $(uptime -p)"
echo "------------------------------------------"
echo "[MEMORY USAGE]"
free -h
echo "------------------------------------------"
echo "[DISK USAGE]"
df -h /
echo "=========================================="
EOF
```

### Langkah 2: Beri Hak Akses Eksekusi & Jalankan

```bash
chmod +x sys-inspect.sh
./sys-inspect.sh
```

**Hasil Expected Output:**
```text
==========================================
      SYSTEM HEALTH INSPECTION REPORT
==========================================
Date/Time : Sat Aug 15 10:30:00 WIB 2026
Hostname  : dev-machine
User      : developer
Uptime    : up 2 hours, 15 minutes
------------------------------------------
[MEMORY USAGE]
               total        used        free
Mem:           7.7Gi       2.1Gi       5.6Gi
------------------------------------------
[DISK USAGE]
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        50G   12G   36G  26% /
==========================================
```

---

## 7. Target & Checklist Capaian Pembelajaran

Gunakan checklist ini untuk memverifikasi pemahaman Anda:

- [ ] Memahami perbedaan hirarki direktori Linux (`/etc`, `/var`, `/proc`, `/home`).
- [ ] Mampu menggunakan perintah navigasi CLI (`ls -lah`, `cd`, `find`, `grep`, `less`).
- [ ] Memahami konsep permission `rwx` (numeric vs symbolic) dan prinsip Least Privilege.
- [ ] Mampu melihat proses berjalan (`ps aux`), memantau RAM/Disk (`free`, `df`), dan mengelola service (`systemctl`, `journalctl`).
- [ ] Berhasil membuat dan menguji shell script inspeksi sistem sederhana.
