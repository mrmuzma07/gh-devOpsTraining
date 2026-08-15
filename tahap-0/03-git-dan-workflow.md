# Modul 03 — Git dan Workflow Kolaborasi

## Tujuan

Menggunakan Git sebagai sumber kebenaran perubahan konfigurasi dan kode. Workflow yang rapi mengurangi risiko perubahan manual, memudahkan review, dan menjadi prasyarat GitOps.

## Materi Inti

### 1. Model mental Git

- Working tree, staging area, commit, branch, tag, dan remote.
- `HEAD`, hash commit, parent commit, dan mengapa commit sebaiknya kecil serta dapat dipahami.
- `.gitignore`, line ending, dan perbedaan file tracked/untracked.

```bash
git status
git diff
git add path/to/file
git diff --cached
git commit -m "jelaskan perubahan secara singkat"
```

### 2. Workflow perubahan

- Buat branch dari branch utama yang terbaru.
- Satu perubahan logis per commit; tulis pesan yang menjelaskan alasan perubahan.
- Push branch, buka pull request, baca hasil CI, tanggapi review, lalu merge sesuai kebijakan repository.
- Bedakan `merge` dan `rebase`; jangan rebase branch bersama tanpa koordinasi.

```bash
git switch -c latihan/readme
# edit file, lalu:
git add README.md
git commit -m "tambahkan catatan latihan"
git fetch origin
git rebase origin/main
git push -u origin latihan/readme
```

### 3. Conflict dan pemulihan

- Baca marker conflict, pilih hasil yang benar, jalankan test, lalu `git add` dan lanjutkan merge/rebase.
- `git restore` untuk membatalkan perubahan working tree, `git restore --staged` untuk mengeluarkan file dari staging.
- `git revert` untuk membatalkan commit yang sudah dibagikan; gunakan `reset` dengan hati-hati hanya pada history lokal.
- `git reflog` membantu menemukan referensi lokal yang hilang.

### 4. Git untuk DevOps

- Simpan manifest dan konfigurasi sebagai kode; pisahkan environment dan hindari secret plaintext.
- Review perubahan permission, image tag, resource limit, endpoint, dan policy sebelum merge.
- Tag release dan dokumentasikan cara rollback.

## Latihan

1. Buat repository latihan berisi README, skrip, dan satu file YAML.
2. Kerjakan dua perubahan pada branch terpisah, lalu buat commit atomik dan pull request.
3. Simulasikan conflict di file teks; selesaikan, jalankan validasi, dan dokumentasikan keputusan.
4. Buat commit yang sengaja salah secara lokal, lalu praktikkan `revert` atau `restore` yang sesuai tanpa menghapus pekerjaan lain.

## Checklist

- [ ] Dapat menjelaskan perbedaan working tree, staging, dan commit.
- [ ] Dapat membuat branch, commit, push, dan pull request.
- [ ] Dapat menyelesaikan conflict serta memverifikasi hasilnya.
- [ ] Dapat memilih `restore`, `revert`, atau `reset` sesuai konteks.
- [ ] Dapat memeriksa perubahan konfigurasi agar tidak membocorkan secret.

## Aturan aman

Jangan menjalankan `git push --force` ke branch bersama tanpa persetujuan. Jangan commit `.env`, token, private key, kubeconfig produksi, atau hasil command yang memuat credential.
