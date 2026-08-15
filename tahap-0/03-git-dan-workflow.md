# Modul 03: Git Fundamental & Workflow Kolaborasi DevOps

> **Target Pembelajaran:** Menguasai model mental Git (Working Tree, Staging, Local/Remote Repo), workflow branching, merge/rebase, penyelesaian conflict, Pull Request, serta praktik GitOps aman.

---

## 1. Mengapa Git Adalah Jantung DevOps & GitOps?

Dalam dunia DevOps modern, Git bukan sekadar tempat menyimpan kode program aplikasi. Git adalah **Single Source of Truth (Satu-satunya Sumber Kebenaran)** untuk seluruh infrastruktur dan konfigurasi sistem.

```
┌─────────────────┐      git push       ┌─────────────────┐      Sync       ┌──────────────────┐
│ Developer Laptop│ ──────────────────> │ GitHub / GitLab │ ──────────────> │ ArgoCD / Cluster │
└─────────────────┘                     └─────────────────┘                 └──────────────────┘
```

**Penerapan Git di DevOps:**
- **Infrastructure as Code (IaC):** File Terraform dan Ansible disimpan dan versi-nya dilacak di Git.
- **GitOps (ArgoCD/Flux):** Perubahan pada manifest Kubernetes di Git akan secara otomatis di-deploy ke cluster produksi.
- **Audit & Compliance:** Setiap perubahan tercatat siapa pelakunya, kapan dilakukan, dan apa alasannya.

---

## 2. Model Mental & 4 Area Utama Git

Untuk menguasai Git, Anda harus memahami bagaimana data berpindah di antara **4 Area Utama**:

```
┌──────────────────┐   git add   ┌──────────────────┐  git commit ┌──────────────────┐  git push  ┌──────────────────┐
│   Working Tree   │ ──────────> │   Staging Area   │ ──────────> │ Local Repository │ ──────────> │ Remote (GitHub)  │
│ (File Diedit)    │ <────────── │ (Index / Draft)  │ <────────── │ (.git/ folder)   │ <────────── │ (Origin / Main)  │
└──────────────────┘ git restore └──────────────────┘ git reset   └──────────────────┘ git fetch  └──────────────────┘
```

### A. Perintah Dasar Manipulasi File

```bash
# 1. Memeriksa status file (Tracked, Untracked, Modified, Staged)
git status

# 2. Melihat perbedaan perubahan kode yang belum di-stage
git diff

# 3. Menambahkan file ke Staging Area (Index)
git add README.md
git add .

# 4. Membuat commit baru dengan pesan semantik
git commit -m "feat: tambahkan modul prasyarat git"

# 5. Melihat riwayat commit secara ringkas & grafis
git log --oneline --graph --decorate -n 5
```

**Contoh Output `git log --oneline`:**
```text
* a1b2c3d (HEAD -> feat/tahap-0, origin/feat/tahap-0) feat: tambahkan modul prasyarat git
* e5f6g7h docs: perbarui panduan instalasi
* i9j0k1l initial commit
```

---

## 3. Workflow Kolaborasi (Branching, Rebase, & Pull Request)

### A. Branching Strategy

Jangan pernah melakukan commit langsung di branch utama (`main` / `master`). Buatlah **Feature Branch** untuk setiap tugas baru.

```bash
# 1. Membuat dan berpindah ke branch baru
git switch -c feat/tambah-materi-linux

# 2. Melihat daftar branch aktif
git branch -a
```

### B. Konvensi Pesan Commit (Semantic Commits)

Format pesan commit yang rapi dan mudah dibaca oleh tim:
- `feat:` Menambahkan fitur atau materi baru.
- `fix:` Memperbaiki bug atau kesalahan penulisan.
- `docs:` Perubahan dokumen/README tanpa mengubah kode.
- `refactor:` Pengorganisasian ulang struktur tanpa mengubah fungsi.

### C. Menjaga History Rapi dengan `git rebase`

```bash
# 1. Ambil pembaruan terbaru dari remote tanpa me-merge
git fetch origin

# 2. Terapkan commit lokal Anda di atas commit terbaru branch main
git rebase origin/main

# 3. Push branch ke remote GitHub
git push -u origin feat/tambah-materi-linux
```

---

## 4. Conflict Resolution & Disaster Recovery

### A. Menyelesaikan Merge Conflict

Conflict terjadi ketika dua orang mengubah **baris file yang sama** pada branch berbeda.

**Tampilan Conflict Marker di File:**
```text
< < < < < < < HEAD (Branch Saat Ini)
server_port: 8080
= = = = = = =
server_port: 9090
> > > > > > > feat/tambah-materi-linux (Commit yang Di-merge)
```

**Langkah Menyelesaikan Conflict:**
1. Buka file yang bermasalah, hapus marker (`< < < < < < <`, `= = = = = = =`, `> > > > > > >`), dan pilih isi kode yang benar.
2. Simpan file, lalu jalankan `git add <nama-file>`.
3. Lanjutkan proses rebase dengan `git rebase --continue` (atau `git commit` jika menggunakan merge).

### B. Penyelamatan & Disaster Recovery (`restore`, `revert`, `reflog`)

```bash
# 1. Pembatalan perubahan di Working Tree (File kembali ke kondisi commit terakhir)
git restore file.txt

# 2. Mengeluarkan file dari Staging Area kembali ke Working Tree
git restore --staged file.txt

# 3. Membatalkan commit yang SUDAH di-push dengan membuat commit pembatalan baru (Aman untuk tim)
git revert <commit-hash>

# 4. Senjata Rahasia SRE: Jurnal Darurat Git untuk menemukan commit yang sengaja/tidaksengaja terhapus
git reflog
```

---

## 5. Best Practices Git untuk DevOps & GitOps

> 🔒 **Security Rules Mandatori:**
> 1. **DILARANG COMMIT SECRET:** Jangan pernah memasukkan `.env`, API Key, Password, Private Key (`id_rsa`), atau Token Service Account ke dalam Git.
> 2. **Gunakan `.gitignore`:** Selalu buat file `.gitignore` di root repositori.
> 3. **Gunakan Branch Protection:** Kunci branch `main` agar hanya bisa diisi melalui Pull Request yang disetujui.

---

## 6. Lab Hands-on: Branching, Conflict Resolution, & Pull Request

Pada latihan ini, Anda akan mensimulasikan pembuatan branch, simulasi conflict, dan penyelesaiannya.

### Langkah 1: Buat Repositori Latihan & Branch

```bash
mkdir -p ~/tahap-0-lab/git-demo
cd ~/tahap-0-lab/git-demo
git init
echo "# Demo Git Ops" > README.md
git add README.md
git commit -m "chore: initial commit"

# Buat feature branch
git switch -c feat/demo-conflict
echo "config_val=100" > config.env
git add config.env
git commit -m "feat: tambahkan config.env"
```

### Langkah 2: Simulasikan Perubahan di Main

```bash
# Kembali ke main dan buat perubahan yang bertentangan
git switch main
echo "config_val=200" > config.env
git add config.env
git commit -m "feat: tambahkan config.env beda nilai di main"
```

### Langkah 3: Lakukan Rebase & Selesaikan Conflict

```bash
# Cobalah rebase branch feature ke main
git switch feat/demo-conflict
git rebase main
# (Git akan menghentikan rebase dan melaporkan CONFLICT!)

# Selesaikan conflict
echo "config_val=200" > config.env
git add config.env
git rebase --continue
```

**Hasil Expected Output:**
```text
Applying: feat: tambahkan config.env
Successfully rebased and updated refs/heads/feat/demo-conflict.
```

---

## 7. Target & Checklist Capaian Pembelajaran

Gunakan checklist ini untuk memverifikasi pemahaman Anda:

- [ ] Memahami 4 area utama Git (Working Tree, Staging Area, Local Repo, Remote Repo).
- [ ] Mampu membuat dan mengelola Feature Branch (`git switch -c`).
- [ ] Menggunakan format pesan commit semantik (`feat:`, `fix:`, `docs:`).
- [ ] Memahami perbedaan `git merge` dan `git rebase`.
- [ ] Mampu membaca marker conflict dan menyelesaikannya secara manual.
- [ ] Menguasai perintah pemulihan darurat (`git restore`, `git revert`, `git reflog`).
- [ ] Memahami larangan keras mengunggah Secret/Credential ke repositori Git.
