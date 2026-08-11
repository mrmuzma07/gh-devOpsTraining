# Modul 03: Panduan Instalasi & Persiapan Environment

> **Target Pembelajaran:** Berhasil menginstal Podman, k3s (atau k3d), dan kubectl di laptop lokal serta memverifikasi bahwa cluster K8s siap digunakan.

---

## 1. Prasyarat Sistem (Hardware & OS)

Sebelum memulai instalasi, pastikan laptop Anda memenuhi spesifikasi minimum berikut:
- **RAM:** Minimum 4 GB (Disarankan 8 GB+).
- **Disk:** Sisa ruang penyimpanan minimum 10 GB.
- **OS Supported:** macOS (Intel/Apple Silicon), Linux (Ubuntu/Debian/Fedora), atau Windows 11 dengan WSL2.

---

## 2. Langkah 1: Instalasi Podman

Podman digunakan sebagai Container Engine lokal untuk memasang, menguji, dan membangun OCI image.

### A. macOS (via Homebrew)
```bash
# 1. Install Podman via brew
brew install podman

# 2. Inisialisasi & jalankan Podman Virtual Machine
podman machine init
podman machine start

# 3. Verifikasi instalasi
podman info
```

### B. Linux (Ubuntu / Debian)
```bash
# Install podman langsung dari package manager
sudo apt update
sudo apt install -y podman

# Verifikasi instalasi
podman info
```

### C. Windows (via WSL2)
Buka terminal WSL2 (Ubuntu):
```bash
sudo apt update
sudo apt install -y podman
podman info
```

---

## 3. Langkah 2: Instalasi `kubectl`

`kubectl` adalah alat CLI resmi untuk berinteraksi dan mengirim perintah ke cluster Kubernetes.

### A. macOS
```bash
brew install kubernetes-cli
```

### B. Linux (Ubuntu / Debian)
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

### C. Verifikasi kubectl
```bash
kubectl version --client
```
*Output harus menampilkan informasi versi client kubectl.*

---

## 4. Langkah 3: Instalasi & Jalankan Cluster `k3s`

Terdapat 2 opsi populer untuk menjalankan k3s di laptop:

### Opsi A: Linux Native (`k3s` official script)
Sangat mudah untuk pengguna Linux:
```bash
# Download & install k3s
curl -sfL https://get.k3s.io | SH_EVAL=true sh -

# Berikan izin akses pada kubeconfig agar bisa dibaca tanpa sudo
mkdir -p ~/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
export KUBECONFIG=~/.kube/config
```

### Opsi B: macOS / Windows / Multi-platform (`k3d` - k3s di dalam Container/Podman)
`k3d` adalah wrapper ringan untuk menjalankan cluster `k3s` di dalam container Docker/Podman.

```bash
# 1. Install k3d
# macOS:
brew install k3d

# Linux / WSL2:
wget -qO- https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# 2. Buat cluster k3s baru dengan k3d
k3d cluster create mini-prod --agents 1

# k3d otomatis mengatur file ~/.kube/config Anda!
```

---

## 5. Verifikasi Cluster (Sanity Check)

Setelah memasang k3s, lakukan verifikasi koneksi cluster dengan perintah berikut:

### 1. Cek Informasi Cluster
```bash
kubectl cluster-info
```
*Output yang benar:*
```text
Kubernetes control plane is running at https://127.0.0.1:6443
CoreDNS is running at https://127.0.0.1:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

### 2. Cek Status Node
```bash
kubectl get nodes
```
*Output yang benar:*
```text
NAME                     STATUS   ROLES                  AGE   VERSION
k3s-node-1 / k3d-mini    Ready    control-plane,master   1m    v1.28.x+k3s1
```

> **Selamat!** Cluster Kubernetes (k3s) Anda kini aktif dan siap menerima deployment pertama Anda.
