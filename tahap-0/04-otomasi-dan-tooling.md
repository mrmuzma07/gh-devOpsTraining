# Modul 04 — Otomasi dan Tooling Dasar

## Tujuan

Membaca format konfigurasi dan membuat otomasi kecil yang dapat diulang. Skill ini dipakai terus saat menulis manifest Kubernetes, pipeline CI, dan runbook.

## Materi Inti

### 1. YAML dan JSON

- Indentasi YAML menggunakan spasi, mapping, list, scalar, quote, dan komentar.
- JSON object, array, string, number, boolean, dan `null`.
- Perbedaan nilai string dan boolean; pahami bahwa parser bisa memperlakukan `yes`, angka, atau tanggal secara berbeda.
- Validasi struktur sebelum menjalankan konfigurasi.

```bash
python3 -m json.tool data.json
# Gunakan parser/validator YAML yang tersedia di environment lab.
```

### 2. Environment variable dan secret

- Environment variable memisahkan konfigurasi dari kode, tetapi bukan tempat aman untuk menampilkan secret.
- Gunakan placeholder pada contoh; redaksi token dari log dan screenshot.
- Bedakan `unset`, string kosong, dan nilai default.

```bash
: "${API_URL:?API_URL wajib diisi}"
printf 'API_URL=%s\n' "$API_URL"
```

### 3. Shell script dan CLI

- Shebang, argument, quote, exit code, `set -euo pipefail`, function, dan cleanup.
- Idempotensi: menjalankan skrip dua kali tidak menghasilkan keadaan yang rusak.
- `--help`, `--dry-run`, output terstruktur, dan logging yang dapat ditelusuri.

### 4. Membaca dokumentasi

- Mulai dari tujuan, prerequisites, versi, contoh minimal, cleanup, dan known limitations.
- Catat asumsi dan versi tool; jangan menganggap command dari macOS, Linux, dan Windows identik.

## Latihan

1. Buat skrip `check-prerequisites.sh` yang memeriksa keberadaan `git`, `curl`, dan satu runtime container.
2. Tambahkan mode `--dry-run` serta exit code berbeda untuk dependency hilang dan input tidak valid.
3. Buat contoh YAML yang mendeskripsikan service sederhana, lalu validasi indentasi dan tipe nilainya.
4. Tulis README singkat yang menjelaskan prerequisites, cara menjalankan, output yang diharapkan, dan cleanup.

## Checklist

- [ ] Dapat membaca YAML/JSON dan menemukan kesalahan indentasi atau tipe.
- [ ] Dapat menulis skrip yang menangani argument dan exit code.
- [ ] Dapat membuat command aman terhadap spasi dan nilai kosong.
- [ ] Dapat membedakan konfigurasi, secret, dan data observability.
- [ ] Dapat menemukan versi tool serta mengikuti dokumentasi resminya.
