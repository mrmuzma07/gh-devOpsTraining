# Minggu 10 — Incident Simulation II (Complex Multi-Source Correlation)

Selamat datang di **Minggu 10: Incident Simulation II**. Modul ini merupakan puncaknya pelatihan SRE / DevOps, di mana Anda dilatih untuk menangani **9 insiden kompleks dunia nyata** yang tidak dapat diselesaikan hanya dengan satu tool, melainkan membutuhkan korelasi multi-sumber: **Mimir (Metrics) + Loki (Logs) + Tempo (Traces) + `kubectl` (Cluster State)**.

---

## 📚 Daftar Modul Pembelajaran

| Modul | Judul Insiden / Topik | Deskripsi Ringkas & Resource | Focus Tools |
|---|---|---|---|
| **[01](./01-metodologi-incident-kompleks.md)** | **Metodologi Incident Kompleks** | Kerangka kerja korelatif 4 sumber observabilitas & 7-Layer Drill-Down. | Mimir, Loki, Tempo, kubectl |
| **[02](./02-incident-cpu-spike.md)** | **CPU Spike & Throttling** | Menangani CPU Throttling, profiling pprof Go/Java, & HPA. | pprof, Mimir, cgroups |
| **[03](./03-incident-memory-leak.md)** | **Memory Leak & Heap Analysis** | Heap dump analysis, LRU Cache unbounded map fix, & deriv() alert. | Go pprof, jmap, Mimir |
| **[04](./04-incident-disk-full.md)** | **Disk Pressure & Full Node** | Penanganan ephemeral-storage overflow, image GC tuning, & logrotate. | crictl, df, du, kubelet |
| **[05](./05-incident-dns-error.md)** | **CoreDNS Failures & Resolution** | Debugging `no such host`, `ndots:5` tuning, & NodeLocal DNSCache. | dig, nslookup, CoreDNS |
| **[06](./06-incident-pvc-full.md)** | **PVC Capacity Exhaustion** | Live Volume Expansion (CSI), Online resizing, & alert linear prediction. | StorageClass, PVC, CSI |
| **[07](./07-incident-network-timeout.md)** | **Network Timeout & CNI Issues** | Troubleshooting packet drop, NetworkPolicy egress/ingress, & tcpdump. | netshoot, tcpdump, netpol |
| **[08](./08-incident-latency-n1-query.md)** | **High Latency N+1 Query** | Deteksi anti-pattern N+1 query via Tempo trace waterfall & code fix. | Tempo, OpenTelemetry, ORM |
| **[09](./09-incident-slow-database.md)** | **Slow Database & Missing Index** | Tuning slow query, Sequential Scan fix (`CREATE INDEX CONCURRENTLY`). | pg_stat_statements, psql |
| **[10](./10-incident-deadlock.md)** | **Database Deadlocks (40P01)** | Penanganan error 40P01, Consistent Lock Ordering, & Retry backoff. | pg_locks, Go SQL Driver |

---

## 🗺️ Matriks Pemetaan Insiden & Sumber Observabilitas

| Insiden # | Mimir (Kapan & Berapa) | Loki (Apa yang Error) | Tempo (Di mana Bottleneck) | `kubectl` (State System) |
|---|---|---|---|---|
| **#1 CPU Spike** | `container_cpu_cfs_throttled_periods_total` | Log request spike / GC log | Span app HTTP handler slow | `top node`, `top pod` |
| **#2 Memory Leak** | `container_memory_working_set_bytes` (upward slope) | `OOMKilled` Exit 137 event | Span heap allocation grow | `describe pod` (Limits) |
| **#3 Disk Full** | `node_filesystem_avail_bytes` < 10% | `ENOSPC: no space left` | Span `kubelet.evictPod` | `describe node` (DiskPressure) |
| **#4 DNS Error** | `coredns_dns_responses_total{rcode!="NOERROR"}` | `dial tcp: lookup ... no such host` | Span DNS resolver timeout | `get pods -n kube-system` |
| **#5 PVC Full** | `kubelet_volume_stats_used_bytes` > 90% | `PANIC: No space left on device` | Span DB Write Failure | `get pvc`, `describe pvc` |
| **#6 Net Timeout** | `http_request_duration_seconds_bucket` (30s) | `Connection timed out` | Span HTTP client hung (No child) | `get netpol`, `exec netshoot` |
| **#7 Latency N+1** | p95/p99 Latency > 5s | Multi-SQL queries per ms | **100+ DB Spans Staircase** | `top pod` (App low CPU) |
| **#8 Slow DB** | `pg_stat_user_tables_seq_scan` > 80% | `LOG: duration: 12000ms` | Single long-duration DB span | `exec postgres` (`psql`) |
| **#9 Deadlock** | `rate(pg_stat_database_deadlocks)` > 0 | `deadlock detected (SQLSTATE 40P01)` | Transaksi span aborted | `psql` (`pg_locks`) |

---

## 🛠️ Folder Manifests SRE

Seluruh manifest YAML untuk mereproduksi ke-9 insiden lab di lingkungan lokal (k3s / minikube) tersedia di direktori `manifests/`:

```text
minggu-10/manifests/
├── 01-cpu-spike-stress.yaml      # Lab #1: CPU Stressing & Throttling
├── 04-disk-fill.yaml             # Lab #3: Ephemeral Storage Overflow
├── 05-dns-broken-config.yaml     # Lab #4: CoreDNS Broken Config & Netpol
├── 06-pvc-lab.yaml               # Lab #5: PVC Resizing & Online Expansion
├── 07-network-timeout-lab.yaml   # Lab #6: NetworkPolicy Isolation Lab
├── 08-n1-app-lab.yaml            # Lab #7: N+1 Tracing App Setup
└── 09-postgres-slow-query-lab.yaml # Lab #8 & #9: PostgreSQL Lab Setup
```

---

## 🎯 Panduan Ujian Evaluasi Mandiri SRE (Self-Assessment)

Sebelum melanjutkan ke **Minggu 11: Reliability Engineering (SLA/SLO/SLI & Error Budgets)**, pastikan Anda dapat menjawab pertanyaan berikut:

1. **Kapan Anda memilih menggunakan `crictl rmi --prune` dibanding `kubectl delete pod`?**
2. **Mengapa `CREATE INDEX CONCURRENTLY` wajib digunakan saat memperbaiki slow query di Production PostgreSQL?**
3. **Bagaimana cara membedakan insiden Network Policy Blocking dengan Port Listener Down tanpa melihat manifest YAML?**
4. **Mengapa `ndots:5` pada `/etc/resolv.conf` di Pod Kubernetes dapat menyebabkan overhead signifikan pada CoreDNS?**

---

**Selamat! Anda telah menyelesaikan detail materi pembelajaran Minggu 10: Incident Simulation II.**
