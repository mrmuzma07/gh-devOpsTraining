# Modul 04: Otomasi, Shell Scripting, & Tooling Dasar

> **Target Pembelajaran:** Memahami format data konfigurasi YAML/JSON, manajemen Environment Variables, prinsip penulisan Bash Script aman (set -euo pipefail), dan metode membaca dokumentasi CLI secara efektif.

---

## 1. Format Konfigurasi: YAML vs JSON

Dalam ekosistem SRE dan Kubernetes, konfigurasi ditulis dalam format **YAML** atau **JSON**. Pemahaman sintaksis yang presisi sangat krusial karena kesalahan 1 spasi saja di YAML bisa menggagalkan deployment.

### A. Perbandingan YAML vs JSON

| Fitur | YAML | JSON |
| :--- | :--- | :--- |
| **Keterbacaan** | Sangat ramah manusia (tanpa kurung siku/kurawal) | Lebih kaku (banyak `{}` dan `[]`) |
| **Indentasi** | Wajib menggunakan **Spasi** (Dilarang Tab!) | Tidak terpengaruh indentasi spasi |
| **Komentar** | Mendukung komentar (`# Komentar`) | TIDAK mendukung komentar |
| **Penggunaan Utama** | Manifest K8s, Ansible, Docker Compose, CI/CD | API Payload, File Konfigurasi Tooling (`package.json`) |

### B. Contoh Perbandingan Kode

**Format JSON (`config.json`):**
```json
{
  "appName": "user-service",
  "port": 8080,
  "enabled": true,
  "tags": ["web", "api"]
}
```

**Format YAML (`config.yaml`):**
```yaml
# Konfigurasi Aplikasi Web
appName: user-service
port: 8080
enabled: true
tags:
  - web
  - api
```

> ⚠️ **Jebakan Sintaksis YAML:**
> - **Boolean Trap:** Kata `yes`, `no`, `true`, `false` tanpa tanda petik bisa diterjemahkan sebagai boolean oleh parser YAML.
> - **Tab Trap:** Jangan gunakan karakter `TAB` untuk indentasi. Gunakan selalu 2 spasi.

---

## 2. Environment Variables & Secret Management

### A. Penggunaan Variable Substitution di Shell

Di skrip DevOps, kita sering membutuhkan nilai bawaan (*default value*) atau validasi variabel lingkungan.

```bash
# 1. Menetapkan nilai default jika variabel PORT belum diset (Default: 8080)
APP_PORT="${PORT:-8080}"
echo "Aplikasi berjalan di port: $APP_PORT"

# 2. Memberikan pesan ERROR dan menghentikan skrip jika DATABASE_URL kosong
: "${DATABASE_URL:?Error: Variables DATABASE_URL wajib diisi!}"
```

### B. Praktik Aman Pengelolaan Secret

1. Buat file `.env.example` yang di-commit ke Git sebagai contoh template:
   ```text
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=app_user
   DB_PASS=YOUR_PASSWORD_HERE
   ```
2. Buat file `.env` asli yang berisi password sebenarnya, dan pastikan `.env` terdaftar di `.gitignore`.

---

## 3. Bash Scripting untuk Otomasi SRE

### A. Boilersuit Aman: `set -euo pipefail`

Setiap shell script produksi di DevOps **WAJIB** mencantumkan opsi keamanan di awal baris skrip:

```bash
#!/usr/bin/env bash
set -euo pipefail
```

**Penjelasan Detail Flag Keamanan:**
- `set -e`: Menghentikan eksekusi skrip seketika jika ada perintah yang menghasilkan exit status selain `0` (error).
- `set -u`: Menghasilkan error jika skrip mencoba membaca variabel yang belum didefinisikan (*unset variable*).
- `set -o pipefail`: Memastikan seluruh rantai pipeline (`cmd1 | cmd2`) gagal jika ada *satu saja* perintah di dalam pipeline yang gagal.

### B. Exit Code (Status Keluar Perintah)

Setiap perintah Linux mengembalikan nilai angka saat selesai:
- `0`: Sukses / Berhasil.
- Non-zero (`1` s/d `255`): Terjadi Error / Gagal.

```bash
# Memeriksa exit status perintah sebelumnya
echo $?
```

### C. Prinsip Idempotensi (Idempotency)

Skrip otomasi yang baik harus bersifat **idempotent**: Menjalankan skrip 1 kali atau 100 kali harus menghasilkan kondisi akhir sistem yang **sama persis**, tanpa menyebabkan error.

*Contoh Salah (Tidak Idempotent):*
`mkdir my_folder` $
ightarrow$ Jalankan kedua kalinya akan error (`File exists`).

*Contoh Benar (Idempotent):*
`mkdir -p my_folder` $
ightarrow$ Aman dijalankan berulang kali.

---

## 4. Strategi Membaca Dokumentasi CLI & Tooling

Seorang SRE handal tidak menghafal seluruh flag perintah CLI. Mereka ahli membaca dokumentasi dengan cepat.

```bash
# 1. Melihat opsi bantuan ringkas
kubectl --help
docker run --help

# 2. Membaca manual lengkap di Linux
man grep
```

---

## 5. Lab Hands-on: Membuat Skrip Prerequisite Checker

Pada latihan ini, Anda akan membuat skrip otomatis untuk mengecek kesiapan alat kerja di komputer Anda.

### Langkah 1: Buat File `check-prereq.sh`

```bash
mkdir -p ~/tahap-0-lab/automation
cd ~/tahap-0-lab/automation

cat << 'EOF' > check-prereq.sh
#!/usr/bin/env bash
set -euo pipefail

echo "===> Checking DevOps Tool Prerequisites <==="

check_tool() {
  local tool_name="$1"
  if command -v "$tool_name" >/dev/null 2>&1; then
    echo "[OK] $tool_name is installed -> $(command -v "$tool_name")"
  else
    echo "[ERROR] $tool_name is NOT installed!"
    return 1
  fi
}

FAILED=0
check_tool "git" || FAILED=1
check_tool "curl" || FAILED=1
check_tool "python3" || FAILED=1

if [ "$FAILED" -eq 1 ]; then
  echo "--------------------------------------------"
  echo "Result: Some required tools are missing!"
  exit 1
else
  echo "--------------------------------------------"
  echo "Result: All prerequisite tools are ready!"
  exit 0
fi
EOF
```

### Langkah 2: Jalankan Script & Uji Exit Status

```bash
chmod +x check-prereq.sh
./check-prereq.sh
echo "Exit Status: $?"
```

**Hasil Expected Output:**
```text
===> Checking DevOps Tool Prerequisites <===
[OK] git is installed -> /usr/bin/git
[OK] curl is installed -> /usr/bin/curl
[OK] python3 is installed -> /usr/bin/python3
--------------------------------------------
Result: All prerequisite tools are ready!
Exit Status: 0
```

---

## 6. Target & Checklist Capaian Pembelajaran

Gunakan checklist ini untuk memverifikasi pemahaman Anda:

- [ ] Memahami perbedaan sintaksis YAML vs JSON serta aturan indentasi 2 spasi.
- [ ] Mampu menetapkan default value environment variable (`${VAR:-default}`).
- [ ] Mengerti alasan wajib mencantumkan `set -euo pipefail` di skrip Bash.
- [ ] Memahami makna Exit Code `0` (Sukses) dan Non-Zero (Error).
- [ ] Memahami prinsip Idempotensi dalam penulisan skrip otomasi.
- [ ] Berhasil membuat dan menguji skrip pengecek prasyarat tools.
