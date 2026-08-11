# Modul 04: Perintah Dasar CLI `kubectl`

> **Target Pembelajaran:** Menguasai 5 perintah dasar `kubectl` paling penting (`get`, `describe`, `logs`, `exec`, `top`) beserta pemahaman cara membaca output dan menggunakannya dalam pekerjaan sehari-hari.

---

## 1. Anatomi Perintah `kubectl`

Format standar penulisan perintah `kubectl` adalah:

```bash
kubectl [action] [resource-type] [resource-name] [flags]
```

**Contoh:**
- `kubectl get pods` *(Lihat semua pod di namespace default)*
- `kubectl describe pod nginx-app` *(Lihat detail pod bernama nginx-app)*
- `kubectl get service -n kube-system` *(Lihat service di namespace kube-system)*

---

## 2. Mendalami 5 Perintah Dasar

### A. `kubectl get` — Melihat Daftar Resource

Gunakan `kubectl get` untuk menampilkan ringkasan daftar resource yang sedang berjalan di cluster.

```bash
# 1. Melihat semua Pod di namespace aktif
kubectl get pods

# 2. Melihat Pod dengan informasi tambahan (IP Address & Node tempat Pod berjalan)
kubectl get pods -o wide

# 3. Melihat resource di Namespace spesifik
kubectl get pods -n kube-system

# 4. Melihat seluruh jenis resource dasar sekaligus
kubectl get all

# 5. Output format YAML/JSON (sangat berguna untuk melihat konfigurasi lengkap)
kubectl get pod <nama-pod> -o yaml
```

**Contoh Output `kubectl get pods -o wide`:**
```text
NAME                     READY   STATUS    RESTARTS   AGE   IP          NODE
nginx-6799fc88d8-x8z2l   1/1     Running   0          5m    10.42.0.12  k3s-node-1
```

---

### B. `kubectl describe` — Memeriksa Detail & Event Log

`kubectl describe` adalah **senjata utama troubleshooting** saat terjadi kendala (misalnya Pod berstatus `CrashLoopBackOff`, `ImagePullBackOff`, atau `Pending`). Perintah ini menampilkan event historis dari Kubernetes.

```bash
# Memeriksa detail Pod
kubectl describe pod <nama-pod>

# Memeriksa detail Deployment
kubectl describe deployment <nama-deployment>
```

**Bagian Penting Output `kubectl describe`:**
1. **Status / State:** Apakah Running, Waiting, atau Terminated.
2. **IP & Node:** Lokasi Pod.
3. **Containers:** State, Last State, Exit Code, Port, Mounts.
4. **Events (Paling Bawah):** Menampilkan aktivitas terkini seperti *Pulling image*, *Successfully assigned*, atau *Failed to pull image*.

---

### C. `kubectl logs` — Membaca Log Aplikasi

Gunakan `kubectl logs` untuk melihat log (*stdout/stderr*) yang dihasilkan oleh aplikasi di dalam container.

```bash
# 1. Melihat log Pod terkini
kubectl logs <nama-pod>

# 2. Stream/Follow log secara realtime (seperti `tail -f`)
kubectl logs -f <nama-pod>

# 3. Menampilkan N baris log terakhir
kubectl logs --tail=50 <nama-pod>

# 4. Jika Pod memiliki lebih dari 1 container (multi-container)
kubectl logs <nama-pod> -c <nama-container>
```

---

### D. `kubectl exec` — Eksekusi Perintah di Dalam Container

Gunakan `kubectl exec` untuk "masuk" ke dalam container atau menjalankan perintah diagnostic di dalamnya (seperti `curl`, `ping`, atau memeriksa file).

```bash
# 1. Masuk ke shell interaktif (bash / sh) di dalam container
kubectl exec -it <nama-pod> -- /bin/sh

# 2. Menjalankan 1 perintah langsung tanpa masuk ke shell
kubectl exec <nama-pod> -- env
kubectl exec <nama-pod> -- curl -I http://localhost
```

> **Catatan:** `--` (double dash) digunakan untuk memisahkan flag `kubectl` dengan perintah yang akan dijalankan di dalam container.

---

### E. `kubectl top` — Memantau Penggunaan CPU & Memory

Perintah ini digunakan untuk melihat secara real-time berapa banyak resource RAM dan CPU yang dikonsumsi oleh Node maupun Pod.

```bash
# 1. Melihat konsumsi resource semua Node
kubectl top nodes

# 2. Melihat konsumsi resource semua Pod di namespace default
kubectl top pods

# 3. Melihat konsumsi resource Pod di semua namespace
kubectl top pods --A
```

**Contoh Output `kubectl top pods`:**
```text
NAME                     CPU(cores)   MEMORY(bytes)
nginx-6799fc88d8-x8z2l   2m           14Mi
```
*(Artinya: 2m = 2 millicores CPU / 0.002 core, 14Mi = 14 Megabytes RAM)*

---

## 3. Cheat Sheet kubectl

| Tugas | Perintah kubectl |
| :--- | :--- |
| **List Pods** | `kubectl get pods` |
| **Detail Pod** | `kubectl describe pod <nama-pod>` |
| **Cek Log** | `kubectl logs -f <nama-pod>` |
| **Masuk Terminal Pod** | `kubectl exec -it <nama-pod> -- /bin/sh` |
| **Cek Penggunaan RAM/CPU** | `kubectl top pods` |
| **Hapus Pod** | `kubectl delete pod <nama-pod>` |
| **Apply Config YAML** | `kubectl apply -f manifest.yaml` |
| **Delete Config YAML** | `kubectl delete -f manifest.yaml` |
