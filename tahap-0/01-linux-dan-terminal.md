# Modul 01 — Linux dan Terminal

## Tujuan

Memahami lingkungan tempat sebagian besar workload, runner CI, dan node Kubernetes berjalan. Fokusnya bukan menghafal semua perintah, melainkan mengetahui cara membaca keadaan sistem dengan aman.

## Materi Inti

### 1. Shell dan filesystem

- Shell, terminal emulator, prompt, `PATH`, dan exit code.
- Direktori penting: `/`, `/etc`, `/var`, `/tmp`, `/home`, `/usr`, `/proc`, dan `/sys`.
- Navigasi serta pencarian: `pwd`, `ls`, `cd`, `find`, `grep`, `less`, `head`, `tail`, dan `man`.
- Redirection dan pipeline: `>`, `>>`, `2>`, `|`, `tee`, serta quote (`'` dan `"`).

```bash
pwd
ls -lah
find . -maxdepth 2 -type f
printf 'status=%s\n' "$?"
```

### 2. File, permission, dan user

- File biasa, direktori, symbolic link, owner, group, dan mode `rwx`.
- `chmod`, `chown`, `umask`, `sudo`, dan prinsip least privilege.
- Perbedaan path relatif dan absolut; hindari `sudo` atau `rm -rf` tanpa memahami targetnya.

```bash
id
ls -l ./script.sh
chmod u+x ./script.sh
```

### 3. Proses dan service

- PID, parent process, foreground/background, signal, dan exit status.
- `ps`, `top` atau `htop`, `pgrep`, `kill`, `jobs`, `fg`, `bg`, dan `systemctl` pada Linux.
- Log service dengan `journalctl`; pahami bahwa nama dan perintah service dapat berbeda antar distro.

```bash
ps aux | head
pgrep -af ssh
free -h
df -h
```

### 4. Package, environment, dan SSH

- Package manager (`apt`, `dnf`, atau `apk`) dan pentingnya versi package.
- Environment variable, file konfigurasi, dan perbedaan shell startup.
- SSH key, `~/.ssh/config`, host verification, port forwarding, dan larangan menyimpan private key di Git.

## Latihan

1. Buat direktori latihan dengan subdirektori `logs`, `scripts`, dan `tmp`.
2. Buat skrip yang menampilkan hostname, user, disk usage, memory, dan lima proses teratas.
3. Jalankan skrip dengan permission yang benar, simpan output menggunakan `tee`, lalu jelaskan exit code-nya.
4. Temukan file log terbaru di environment lab dan tulis tiga observasi tanpa mengubah file tersebut.

## Checklist

- [ ] Dapat menjelaskan permission `rwxr-x---` dan owner file.
- [ ] Dapat membedakan proses, thread, service, dan container secara umum.
- [ ] Dapat membaca disk/memory/process dengan perintah inspeksi.
- [ ] Dapat memakai SSH tanpa membagikan private key.
- [ ] Dapat mencari dokumentasi command melalui `man` atau `--help`.

## Catatan lintas platform

Tahap 1 memakai Linux container dan Kubernetes. Pengguna macOS atau Windows boleh belajar dari host masing-masing, tetapi sebaiknya menjalankan latihan Linux di OrbStack, Docker, WSL2, atau VM agar perilakunya mendekati node produksi.
