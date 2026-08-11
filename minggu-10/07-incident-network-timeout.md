# Incident #7 — Network Connection Timeouts & CNI Issues

> **Satu kalimat:** Ketika terjadi kegagalan jaringan internal Kubernetes, **request antar-service menggantung (hang) selama puluhan detik sebelum akhirnya gagal dengan error `Connection Timeout`**, yang umumnya disebabkan oleh miskonfigurasi NetworkPolicy, masalah Pod CNI Plugin, atau tabel Overlay Network (VXLAN/Calico/Flannel) yang korup.

Masalah jaringan (networking) adalah salah satu tipe insiden yang paling membingungkan bagi SRE pemula karena **semua Pod tampak `Running 1/1`**, tetapi aplikasi di dalamnya tidak dapat saling berkomunikasi.

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Memahami alur komunikasi Pod-to-Pod (CNI, kube-proxy iptables/IPVS, Overlay Network).
2. Membedakan error **Connection Refused** (aplikasi down/port salah) vs **Connection Timeout** (network packet drop/blocked).
3. Melakukan debugging low-level jaringan menggunakan tool `tcpdump`, `traceroute`, `curl`, dan `netstat` di dalam Pod.
4. Mendiagnosis dan memperbaiki masalah **NetworkPolicy** Egress/Ingress blocking.
5. Menganalisis trace komunikasi antar-service yang terhenti menggunakan **Tempo**.

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari sisi Log Aplikasi
- Error khas paket hilang/dibuang (dropped):
  - Go: `Connection to auth-service:8080 timed out after 30s (Client.Timeout exceeded while awaiting headers)`
  - Java: `java.net.SocketTimeoutException: Read timed out`
  - Python/Curl: `curl: (28) Failed to connect to order-service port 8080 after 5001 ms: Couldn't connect to server`

> **Beda Kunci:**
> - `Connection Refused` = Paket sampai ke tujuan, tapi tidak ada process yang listen di port tersebut (atau Port salah).
> - `Connection Timeout` = Paket **DIBUANG (DROPPED)** di tengah jalan oleh Firewall / NetworkPolicy / IP Table rules.

### 1.2 Dari sisi Distributed Tracing (Tempo)
- Span HTTP Client di `frontend-service` berdurasi tepat 30.00s (Default timeout limit).
- Span child untuk `backend-service` **TIDAK PERNAH MUNCUL** di Tempo dashboard (karena paket request HTTP tidak pernah sampai ke target).

```text
[frontend-service] GET /checkout  ===============================================> (30.01s - ERROR)
    └── [http.client] Outbound request to backend-service:8080 ==================> (30.00s - TIMEOUT)
        (No child spans from backend-service)
```

---

## 2. 🔍 Investigasi Step-by-Step (Networking Diagnostic Flow)

```
[ Pod A ] ───(1. Pod Local)───> [ veth pair ] ───(2. CNI Bridge/VXLAN)───> [ Kube-Proxy / Iptables ] ───(3. Target Pod B)
```

### Step 1 — Cek IP Pod & Status Target Application
Pastikan Pod target benar-benar memiliki IP dan port aplikasi listen dengan baik:
```bash
$ kubectl get pods -o wide -n production
NAME                  READY   STATUS    IP           NODE
frontend-service-xxx  1/1     Running   10.42.1.15   laptop-node-1
backend-service-yyy   1/1     Running   10.42.2.40   laptop-node-1

# Cek apakah Port aplikasi terdaftar di target Pod
$ kubectl exec backend-service-yyy -n production -- netstat -tlpn
Active Internet connections (only servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name    
tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN      1/main
```

### Step 2 — Tes Connectivity Menggunakan Debug Container (ephemeral container)
Gunakan `nicolaka/netshoot` yang kaya akan tool jaringan:

```bash
# Attach netshoot ke namespace Pod frontend
$ kubectl debug -it frontend-service-xxx -n production --image=nicolaka/netshoot -- bash

# Test 1: Ping Target IP (Layer 3 Check)
netshoot:~# ping -c 3 10.42.2.40
PING 10.42.2.40 (10.42.2.40) 56(84) bytes of data.
--- 10.42.2.40 ping statistics ---
3 packets transmitted, 0 received, 100% packet loss!   <-- PACKET DROPPED!

# Test 2: Netcat Port Test (Layer 4 Check)
netshoot:~# nc -zv -w 5 10.42.2.40 8080
nc: connect to 10.42.2.40 port 8080 (tcp) timed out    <-- TIMEOUT!

# Test 3: Traceroute untuk melihat lokasi titik drop paket
netshoot:~# traceroute -n 10.42.2.40
traceroute to 10.42.2.40 (10.42.2.40), 30 hops max, 60 byte packets
 1  * * *
 2  * * *  (Stuck total di Hop 1)
```

---

## 3. 🎯 Root Cause Analysis & Skenario Umum

### Skenario A — NetworkPolicy Menghalangi Trafik (Ingress / Egress Block)
- **Penyebab:** Ada `NetworkPolicy` yang di-apply di namespace, secara eksplisit atau implisit memblokir Egress dari `frontend-service` atau Ingress menuju `backend-service`.

#### Cara Verifikasi NetworkPolicy:
```bash
$ kubectl get networkpolicy -n production
NAME                   POD-SELECTOR          AGE
deny-all-ingress       <none>                12d
allow-database-only    app=backend-service   2d
```
Inspect policy `allow-database-only`:
```bash
$ kubectl get netpol allow-database-only -n production -o yaml
```
Jika policy menentukan `ingress.from.podSelector.matchLabels.role: db-client`, namun `frontend-service` tidak memiliki label `role: db-client`, maka paket akan **di-drop secara silent oleh kernel iptables**.

### Skenario B — CNI Plugin DaemonSet Crashed / State Stale
- **Penyebab:** Pod CNI (Flannel, Calico, Cilium, atau K3s Flannel-VXLAN) di node mengalami crash atau kehilangan rute interface `flannel.1` / `calico-veth`.
- **Gejala:** Komunikasi Pod dalam 1 Node lancar, tetapi komunikasi Cross-Node mengalami Timeout.

### Skenario C — Kube-Proxy Iptables Rules Out-of-Sync
- **Penyebab:** `kube-proxy` gagal mengupdate tabel routing Service ClusterIP di iptables Node.
- **Gejala:** Memanggil Pod via **Pod IP langsung BERHASIL**, tetapi memanggil via **Service Name / ClusterIP GAGAL Timeout**.

---

## 4. 🛠️ Mitigasi & Langkah Perbaikan

### Tahap 1 — Memperbaiki Miskonfigurasi NetworkPolicy
Jika terbukti NetworkPolicy yang memblokir, perbarui manifest agar mengizinkan trafik yang sah:

File Fix: `manifests/07-network-policy-fix.yaml`
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend-to-backend
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: backend-service
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend-service  # Mengizinkan Pod frontend masuk
    ports:
    - protocol: TCP
      port: 8080
```

Apply perbaikan:
```bash
$ kubectl apply -f manifests/07-network-policy-fix.yaml
networkpolicy.networking.k8s.io/allow-frontend-to-backend created
```

### Tahap 2 — Perbaikan CNI / Kube-Proxy Stale Rules
Jika CNI Plugin atau Kube-Proxy bermasalah:

```bash
# 1. Restart Pod kube-proxy untuk meregenerasi iptables/IPVS rules
$ kubectl rollout restart daemonset kube-proxy -n kube-system

# 2. Jika menggunakan Flannel/Calico di K3s, restart CNI DaemonSet
$ kubectl rollout restart daemonset svclb-traefik -n kube-system (atau CNI Pods)

# 3. Flushing stale iptables manual di host (Jika terdesak di laptop SRE)
$ sudo iptables -F && sudo systemctl restart k3s
```

---

## 5. 🧪 Lab Hands-On: Troubleshooting Network Timeout

### Step 1 — Deploy Environment Lab
Deploy Frontend, Backend, dan NetworkPolicy pembatas.

File: `minggu-10/manifests/07-network-timeout-lab.yaml`
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: net-lab
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: net-lab
  labels:
    app: backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: web
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: net-lab
  labels:
    app: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: client
        image: alpine:3.19
        command: ["sleep", "3600"]
---
# Policy jahat: Isolasi penuh backend dari Pod mana pun!
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: isolate-backend
  namespace: net-lab
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
  - Ingress
  # Empty Ingress = Deny All Ingress Traffic!
```

### Step 2 — Simulasikan Error Timeout
```bash
$ kubectl apply -f manifests/07-network-timeout-lab.yaml

# Coba panggil backend dari frontend Pod
$ kubectl exec -it deployment/frontend -n net-lab -- wget --timeout=3 -qO- http://backend.net-lab.svc.cluster.local
wget: download timed out   <-- ERROR REPRODUCED!
```

### Step 3 — Investigasi dengan Tcpdump
Buka terminal debug dan jalankan `tcpdump` di backend Pod saat frontend mencoba memanggil:

```bash
# Attach debug container di backend Pod
$ kubectl debug -it deployment/backend -n net-lab --image=nicolaka/netshoot -- tcpdump -nn -i any port 80

# Hasil tcpdump saat frontend wget:
# (Kosong! Tidak ada SYN packet yang sampai sama sekali di interface backend container -> Bukti ter-drop di level host NetPol/iptables!)
```

### Step 4 — Fix & Verifikasi
Hapus policy isolasi:
```bash
$ kubectl delete netpol isolate-backend -n net-lab
networkpolicy "isolate-backend" deleted

# Uji kembali
$ kubectl exec -it deployment/frontend -n net-lab -- wget --timeout=3 -qO- http://backend.net-lab.svc.cluster.local
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title> ...  <-- SUCCESS!
```

---

## 6. 📋 Cheat Sheet Diagnosis Network Timeout

```text
DIAGNOSIS STEP                   | COMMAND / TOOL                                | KETERANGAN
---------------------------------|-----------------------------------------------|-----------------------------------------
1. Check App Listen Status       | kubectl exec <pod> -- netstat -tlpn           | Pastikan port terbuka 0.0.0.0
2. Check Network Policies        | kubectl get netpol -n <ns>                    | Periksa apakah ada isolasi ingress/egress
3. Check Service ClusterIP       | kubectl get svc -n <ns>                       | Pastikan Selector match dengan Pod Label
4. Packet Capture Live           | kubectl debug <pod> --image=netshoot -- tcpdump| Inspect apakah paket SYN masuk/keluar
5. Kube-proxy Health             | kubectl logs -n kube-system -l k8s-app=kube-proxy | Periksa error synching iptables/IPVS
```

---

**Lanjut ke insiden berikutnya:** [08-incident-latency-n1-query.md](./08-incident-latency-n1-query.md) — mengidentifikasi dan menangani degradasi performa akibat masalah N+1 Database Query menggunakan OpenTelemetry & Tempo Tracing.
