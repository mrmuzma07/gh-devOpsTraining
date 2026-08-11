# Incident #5 — CoreDNS Failures & Cluster DNS Resolution Errors

> **Satu kalimat:** Ketika CoreDNS mati, overload, atau salah konfigurasi, **seluruh Pod tidak bisa lagi menemukan Service/Database lewat hostname** (`postgres.default.svc.cluster.local`) dan muncul error `dial tcp: lookup postgres: no such host`.

Di Kubernetes, Pod jarang menggunakan IP Address langsung karena IP Pod bisa berubah sewaktu-waktu. Pod menggunakan **FQDN (Fully Qualified Domain Name)** seperti `payment-db.prod.svc.cluster.local`. Semua request resolver ini ditangani oleh **CoreDNS**. Jika CoreDNS bermasalah, aplikasi akan tampak "mati total" padahal container dan Pod-nya sebenarnya dalam keadaan `Running`.

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Memahami arsitektur DNS internal Kubernetes (kube-dns/CoreDNS) dan alur `ndots:5`.
2. Mendiagnosis error DNS (`no such host`, `i/o timeout`, `SERVFAIL`) menggunakan `dig`, `nslookup`, dan `dnstap`.
3. Memperbaiki masalah CoreDNS OOMKilled, High Latency, Upstream Timeout, dan ConfigMap syntax error.
4. Menerapkan solusi skalabilitas DNS menggunakan **NodeLocal DNSCache** dan `autopath`.
5. Memonitor metrics CoreDNS (`coredns_dns_request_duration_seconds_bucket`, `coredns_dns_responses_total`).

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari sisi Aplikasi (Logs)
- Log aplikasi Microservice memuntahkan exception:
  - Go: `dial tcp: lookup auth-service on 10.96.0.10:53: no such host`
  - Java/Spring: `java.net.UnknownHostException: payment-db.prod.svc.cluster.local`
  - Node.js: `Error: getaddrinfo ENOTFOUND redis-cache`
  - Python: `urllib3.exceptions.MaxRetryError: Failed to establish a new connection: [Errno -3] Temporary failure in name resolution`

### 1.2 Dari sisi Network / HTTP
- HTTP 500 / 502 Bad Gateway di Gateway Ingress.
- Latency request melonjak secara konsisten (+5 detik per request) akibat DNS lookup retry timeout.

### 1.3 Dari sisi Mimir & Prometheus Rules
- Alert: `CoreDNSDown` atau `CoreDNSLatencyHigh`.
- Metric PromQL:
```promql
# Rate DNS Request Error (SERVFAIL, FORMERR, NXDOMAIN)
sum(rate(coredns_dns_responses_total{rcode!="NOERROR"}[5m])) 
/ sum(rate(coredns_dns_responses_total[5m])) * 100
# Hasil: 48.5% request gagal!
```

---

## 2. 🧠 Arsitektur DNS Kubernetes & Masalah `ndots:5`

Sebelum troubleshooting, pahami bagaimana Pod mencari nama domain:

```
[Pod: order-service] 
  │  Request: lookup "payment-service"
  ▼
[Read /etc/resolv.conf inside Pod]
  nameserver 10.96.0.10
  search default.svc.cluster.local svc.cluster.local cluster.local
  options ndots:5
  │
  ├─ 1. Query: "payment-service.default.svc.cluster.local" (Hit!)
  ├─ 2. Query untuk domain luar (misal: "api.stripe.com"):
  │     - Stripe.com cuma ada 1 dot (< 5 dots).
  │     - Kubelet memaksa append search domains DULU:
  │       a) api.stripe.com.default.svc.cluster.local -> NXDOMAIN (Fail)
  │       b) api.stripe.com.svc.cluster.local -> NXDOMAIN (Fail)
  │       c) api.stripe.com.cluster.local -> NXDOMAIN (Fail)
  │       d) api.stripe.com -> SUCCESS (Hit ke Upstream DNS)
  │     * Efek: 1 request external domain memicu 4x DNS Queries ke CoreDNS!
```

Jika ada 1000 Pod melakukan request luar, CoreDNS dihantam **4000 QPS** seketika!

---

## 3. 🔍 Investigasi Step-by-Step

### Step 1 — Verifikasi Status Pod CoreDNS
CoreDNS berjalan di namespace `kube-system`:
```bash
$ kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide
NAME                       READY   STATUS             RESTARTS   AGE
coredns-6743829-x8291      0/1     CrashLoopBackOff   12         1h
coredns-6743829-p210a      1/1     Running            0          1h
```

> **Temuan 1:** 1 dari 2 replica CoreDNS CrashLoopBackOff! Ini menjelaskan mengapa sebagian request berhasil dan sebagian timeout (load balancing ke Pod mati).

### Step 2 — Inspect Log CoreDNS
```bash
$ kubectl logs -n kube-system -l k8s-app=kube-dns --tail=100
# Log Pod 1:
2026-08-11T01:20:15Z [FATAL] plugin/errors: /etc/coredns/Corefile:14 - Syntax error: unexpected token "forward ."
# Log Pod 2 (Running):
2026-08-11T01:22:00Z [WARNING] plugin/health: Health check failed for upstream 8.8.8.8:53: i/o timeout
```

### Step 3 — Uji Resolution dari dalam Pod Diagnostic
Deploy Pod临时 `dnsutils` untuk melakukan pengujian manual:
```bash
$ kubectl run dnsutils --image=registry.k8s.io/e2e-test-images/jessie-dnsutils:1.3 --restart=Never -- sleep 3600
pod/dnsutils created

$ kubectl exec -i -t dnsutils -- nslookup postgres.default.svc.cluster.local
Server:		10.96.0.10
Address:	10.96.0.10#53

Name:	postgres.default.svc.cluster.local
Address: 10.96.0.155   # SUCCESS!

$ kubectl exec -i -t dnsutils -- nslookup google.com
;; connection timed out; no servers could be reached   # FAIL! External domain timeout!
```

---

## 4. 🎯 Root Cause Analysis

### Skenario A — CoreDNS CPU Throttling / Resource Starvation
- **Penyebab:** CPU Limits pada Deployment CoreDNS terlalu kecil (misal `cpu: 100m`). Saat ada spike traffic, CPU ter-throttle oleh CFS quota kernel, membuat latency response DNS membengkak dari 2ms ke 5000ms.
- **Gejala:** Metric `coredns_dns_request_duration_seconds` menunjukkan p99 latency > 2s.

### Skenario B — Upstream Forwarder Timeout (ISP / Corporate Network Issue)
- **Penyebab:** Corefile dikonfigurasi menembus `forward . 8.8.8.8` atau `/etc/resolv.conf` milik Host Node. Jika Host Node kehilangan akses internet atau blocking UDP port 53 out-bound, resolusi domain external gagal (`NXDOMAIN` atau `Timeout`).

### Skenario C — ConfigMap `Corefile` Syntax Error
- **Penyebab:** Perubahan manual pada ConfigMap `coredns` di `kube-system` yang salah format. Saat CoreDNS Pod di-restart, Corefile gagal di-parse dan container crash.

### Skenario D — UDP Port 53 Blocked oleh NetworkPolicy
- **Penyebab:** Aplikasi dideploy dengan `NetworkPolicy` strict `Egress`, tetapi lupa menambahkan izin komunikasi UDP/TCP port 53 ke namespace `kube-system` / CoreDNS ClusterIP.

---

## 5. 🛠️ Mitigasi & Perbaikan (Bertahap)

### Tahap 1 — Pertolongan Darurat (Quick Fix)

```bash
# 1. Scale Up Replicas CoreDNS (Jika CPU/Memory Overload)
$ kubectl scale deployment coredns -n kube-system --replicas=4

# 2. Hapus Pod CoreDNS yang Crash/Hang untuk memicu restart dari ConfigMap yang benar
$ kubectl rollout restart deployment coredns -n kube-system

# 3. Bypass DNS untuk domain eksternal kritis sementara via hostAliases di Deployment App
```
Contoh `hostAliases` sementara di Deployment aplikasi:
```yaml
spec:
  template:
    spec:
      hostAliases:
      - ip: "142.250.190.46"
        hostnames:
        - "api.stripe.com"
```

### Tahap 2 — Perbaikan Konfigurasi Corefile

Edit ConfigMap CoreDNS:
```bash
$ kubectl edit configmap coredns -n kube-system
```
Pastikan `Corefile` memiliki format standar yang valid:
```corefile
.:53 {
    errors
    health {
       lameduck 5s
    }
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }
    prometheus :9153
    forward . 1.1.1.1 8.8.8.8 {
       max_concurrent 1000
    }
    cache 30
    loop
    reload
    loadbalance
}
```

---

## 6. 🛡️ Pencegahan Jangka Panjang & Best Practices

### 6.1 Implementasi NodeLocal DNSCache
`NodeLocal DNSCache` menjalankan DaemonSet CoreDNS ringan di setiap Node sebagai IP Loopback (`169.254.20.10`).
- **Keuntungan:**
  - Mengubah request UDP/TCP cross-node menjadi local node lookup (latency < 1ms).
  - Menghindari iptables connection tracking overhead (conntrack table full bug).
  - Membantu caching agresif sehingga mengurangi beban cluster-wide CoreDNS.

```yaml
# manifest snippet NodeLocal DNSCache
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: node-local-dns
  namespace: kube-system
spec:
  template:
    spec:
      hostNetwork: true
      dnsPolicy: Default
      containers:
      - name: node-cache
        image: registry.k8s.io/dns/k8s-dns-node-cache:1.22.28
        resources:
          requests:
            cpu: 25m
            memory: 5 Mi
```

### 6.2 Tuning `dnsConfig` di Level Pod (Mengatasi `ndots:5`)
Untuk Pod microservice yang sering memanggil API luar, turunkan `ndots` menjadi `2` atau `1`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: payment-processor
spec:
  template:
    spec:
      dnsConfig:
        options:
          - name: ndots
            value: "2"
          - name: edns0
```

### 6.3 HPA (Horizontal Pod Autoscaler) untuk CoreDNS
Gunakan `cluster-proportional-autoscaler` agar jumlah Pod CoreDNS bertambah otomatis mengikuti jumlah Node & Core di Cluster.

---

## 7. 🧪 Lab Hands-On: Simulasi CoreDNS Failure

### Step 1 — Break CoreDNS ConfigMap
Simulasikan kerusakan syntax Corefile dengan mengapply ConfigMap yang salah.

File: `minggu-10/manifests/05-dns-broken-config.yaml`
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns-broken
  namespace: default
data:
  Corefile: |
    .:53 {
        errors
        invalid_plugin_name_xyz {
           foo bar
        }
    }
```

### Step 2 — Simulasikan App NetworkPolicy blocking DNS
File: `minggu-10/manifests/05-app-blocked-dns.yaml`
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-dns-egress
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: IsolatedApp
  policyTypes:
  - Egress
  egress:
  # Lupa membuka port 53 UDP/TCP ke CoreDNS!
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
    ports:
    - protocol: TCP
      port: 80
```

### Step 3 — Verifikasi & Recovery
1. Deploy `dnsutils` Pod dengan label `app: IsolatedApp`.
2. Jalankan `nslookup kubernetes.default` -> Terjadi `i/o timeout` karena terhadang `NetworkPolicy`.
3. Hapus NetworkPolicy dan perbaiki ConfigMap CoreDNS.

---

## 8. 📋 Cheat Sheet Troubleshooting DNS

```text
ERROR / SYMPTOM                      | PENYEBAB UTAMA                  | PERBAIKAN / SOLUSI
-------------------------------------|---------------------------------|---------------------------------------------
dial tcp: lookup ... no such host    | Domain tak ditemukan / ndots    | Cek FQDN, atur dnsConfig ndots:2
connection timed out; no servers     | NetworkPolicy / CoreDNS Down   | Cek CoreDNS pod, izinkan Egress Port 53 UDP
coredns CrashLoopBackOff             | Syntax error di Corefile       | Rollback ConfigMap coredns
Latency High (> 2000ms)              | CoreDNS CPU Throttled / Conntrack| Hapus CPU limit / Pasang NodeLocal DNSCache
SERVFAIL                             | Upstream DNS Timeout / Rate limit| Ganti upstream DNS (1.1.1.1 / 8.8.8.8)
```

---

**Lanjut ke insiden berikutnya:** [06-incident-pvc-full.md](./06-incident-pvc-full.md) — saat storage PersistentVolume 100% penuh dan cara melakukakan online volume expansion.
