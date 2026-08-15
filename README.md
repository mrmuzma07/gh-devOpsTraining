# SRE dan DevOps Training

Kurikulum praktik selama 24 minggu untuk membangun, mengamati, mengamankan,
dan mengoperasikan platform aplikasi berbasis Kubernetes. Materi dimulai dari
konsep container dan Kubernetes, kemudian berkembang ke GitOps, observability,
incident response, reliability engineering, high availability, security,
autoscaling, backup, dan disaster recovery.

## Tujuan Pembelajaran

Pada akhir kurikulum, peserta diharapkan mampu:

- Menjalankan workload aplikasi pada Kubernetes secara deklaratif.
- Mengelola delivery aplikasi dengan CI/CD dan GitOps.
- Menggunakan metrics, logs, dan traces untuk memahami kondisi sistem.
- Menangani insiden dengan alur investigasi yang terukur.
- Menetapkan SLI, SLO, error budget, runbook, dan postmortem.
- Mendesain platform yang high available, aman, scalable, dan dapat dipulihkan.
- Menguji perubahan dan kegagalan secara aman di environment lab.

## Peta Kurikulum 24 Minggu

> **Untuk pemula:** selesaikan **Tahap 0** terlebih dahulu. Tahap ini bukan bagian dari hitungan 24 minggu, tetapi menjadi prasyarat agar materi Tahap 1 dapat diikuti dengan nyaman.

### Tahap 0: Prasyarat SRE dan DevOps

Tahap 0 membangun fondasi yang sering diasumsikan sudah dikuasai: Linux dan terminal, networking, Git, otomasi CLI, serta metode troubleshooting. Materi lengkap dan checklist kelulusan tersedia di [direktori Tahap 0](./tahap-0/README.md).

| Modul | Fokus | Materi |
| --- | --- | --- |
| 1 | Linux dan Terminal | Shell, filesystem, permission, proses, service, package manager, dan SSH | [Modul Linux](./tahap-0/01-linux-dan-terminal.md) |
| 2 | Networking Dasar | IP, DNS, port, TCP/UDP, HTTP/TLS, routing, proxy, dan tools diagnosis | [Modul Networking](./tahap-0/02-networking-dasar.md) |
| 3 | Git dan Workflow | Repository, commit, branch, merge/rebase, conflict, remote, pull request, dan recovery | [Modul Git](./tahap-0/03-git-dan-workflow.md) |
| 4 | Otomasi dan Tooling | YAML, JSON, environment variable, shell script, CLI, dan membaca dokumentasi | [Modul Otomasi](./tahap-0/04-otomasi-dan-tooling.md) |
| 5 | Praktik Troubleshooting | Hipotesis, observasi, mitigasi, rollback, runbook, dan postmortem | [Modul Troubleshooting](./tahap-0/05-praktik-troubleshooting.md) |

**Acceptance gate:** sebelum lanjut, pastikan Anda dapat menggunakan terminal Linux, mendiagnosis koneksi HTTP sederhana, membuat dan memulihkan perubahan Git, membaca YAML/JSON, serta menulis langkah troubleshooting yang dapat diulang.

### Tahap 1: Fondasi Platform (Minggu 1-4)

| Minggu | Fokus | Kemampuan dan artefak utama | Materi |
| --- | --- | --- | --- |
| 1 | Container dan Kubernetes | Container, Pod, Deployment, Service, Ingress, dan manifest dasar | [Minggu 1](./minggu-01/README.md) |
| 2 | Kubernetes Workload | Workload controller, ConfigMap, Secret, PVC, probes, dan Go App | [Minggu 2](./minggu-02/README.md) |
| 3 | Helm | Chart, template, values, release, upgrade, dan rollback | [Minggu 3](./minggu-03/README.md) |
| 4 | GitLab CI dan GitOps | GitLab Runner, registry, ArgoCD, auto-sync, dan self-healing | [Minggu 4](./minggu-04/README.md) |

### Tahap 2: Observability dan Respons Insiden (Minggu 5-12)

| Minggu | Fokus | Kemampuan dan artefak utama | Materi |
| --- | --- | --- | --- |
| 5 | Metrics | Prometheus format, PromQL, Alloy, Mimir, dan dashboard Grafana | [Minggu 5](./minggu-05/README.md) |
| 6 | Logging | Structured logging, Alloy, Loki, LogQL, dan investigasi log | [Minggu 6](./minggu-06/README.md) |
| 7 | Distributed Tracing | OpenTelemetry, Tempo, TraceQL, span, dan context propagation | [Minggu 7](./minggu-07/README.md) |
| 8 | Dashboard dan Alert | Golden Signals, RED/USE, alert rule, routing, dan inhibition | [Minggu 8](./minggu-08/README.md) |
| 9 | Incident Simulation I | CrashLoopBackOff, OOMKilled, Pending, ImagePullBackOff, dan FailedMount | [Minggu 9](./minggu-09/README.md) |
| 10 | Incident Simulation II | CPU, memory, disk, DNS, network, latency, database, dan deadlock | [Minggu 10](./minggu-10/README.md) |
| 11 | Reliability Engineering | SLI, SLO, SLA, error budget, capacity planning, runbook, dan postmortem | [Minggu 11](./minggu-11/README.md) |
| 12 | Production Simulation | Integrasi delivery, load test, alert, incident, recovery, dan postmortem | [Minggu 12](./minggu-12/README.md) |

### Tahap 3: Kapabilitas Produksi Lanjutan (Minggu 13-24)

| Minggu | Fokus | Kemampuan dan artefak utama | Materi |
| --- | --- | --- | --- |
| 13 | High Availability | Multi-node k3s, etcd quorum, PDB, node maintenance, dan failover | [Minggu 13](./minggu-13/README.md) |
| 14 | Infrastructure as Code | OpenTofu, Ansible, state management, provisioning, dan drift detection | [Minggu 14](./minggu-14/README.md) |
| 15 | Security Baseline | RBAC, PSA, Kyverno, cert-manager, TLS, dan Trivy | [Minggu 15](./minggu-15/README.md) |
| 16 | Networking | Cilium, eBPF, Hubble, CoreDNS, NetworkPolicy, dan egress gateway | [Minggu 16](./minggu-16/README.md) |
| 17 | Autoscaling | HPA, VPA, KEDA, Cluster Autoscaler, PDB, dan cost optimization | [Minggu 17](./minggu-17/README.md) |
| 18 | Backup dan Disaster Recovery | RPO, RTO, Velero, PgBackRest, 3-2-1 backup, dan DR drill | [Minggu 18](./minggu-18/README.md) |
| 19 | Progressive Delivery | Argo Rollouts, canary, blue-green, dan analysis | [Roadmap lanjutan](./README-Lanjutan.md) |
| 20 | Supply Chain Security | SBOM, Trivy, Syft, Grype, Cosign, signature, dan Renovate | [Roadmap lanjutan](./README-Lanjutan.md) |
| 21 | Cost dan Capacity | OpenCost, rightsizing, forecast, budget, dan capacity model | [Roadmap lanjutan](./README-Lanjutan.md) |
| 22 | Chaos Engineering | Chaos Mesh, hypothesis, blast radius, eksperimen, dan laporan | [Roadmap lanjutan](./README-Lanjutan.md) |
| 23 | Governance dan Upgrade | ArgoCD AppProject, ownership, version pinning, dan compatibility matrix | [Roadmap lanjutan](./README-Lanjutan.md) |
| 24 | Final Game Day | Integrasi semua capability, assessment, dan final postmortem | [Roadmap lanjutan](./README-Lanjutan.md) |

> Materi detail dan direktori praktik yang tersedia di repository saat ini adalah
> Minggu 1-18. Minggu 19-24 sudah didefinisikan sebagai roadmap di
> [`README-Lanjutan.md`](./README-Lanjutan.md) dan dapat dikembangkan menjadi
> modul praktik berikutnya.

## Cara Menggunakan Repository

### 1. Ikuti urutan minggu

Mulai dari Minggu 1 dan lanjutkan secara berurutan. Setiap minggu biasanya
memiliki pola berikut:

1. Baca `README.md` pada direktori minggu tersebut.
2. Pelajari modul teori bernomor paling kecil terlebih dahulu.
3. Siapkan prasyarat yang disebutkan di bagian prerequisites.
4. Jalankan lab menggunakan manifest, source code, atau script yang tersedia.
5. Amati hasilnya dengan perintah CLI dan observability stack.
6. Kerjakan checklist kelulusan sebelum lanjut ke minggu berikutnya.

Jangan menganggap materi selesai hanya karena manifest berhasil di-apply.
Simpan bukti hasil praktik, pahami alasan setiap konfigurasi, dan catat masalah
yang ditemukan.

### 2. Siapkan environment lab

Tool inti yang digunakan sepanjang kurikulum meliputi:

- Linux shell dan utilitas CLI dasar untuk menjalankan latihan serta diagnosis.
- Git untuk version control, code review, dan GitOps.
- `curl` serta utilitas DNS/socket untuk menguji konektivitas.
- `kubectl` untuk berinteraksi dengan cluster.
- `k3s` atau `k3d` untuk cluster Kubernetes lokal.
- OrbStack + Docker CLI untuk membuat dan menjalankan image secara lokal di macOS; gunakan runtime setara pada platform lain.
- `helm` untuk package management dan instalasi komponen cluster.
- GitLab dan ArgoCD untuk CI/CD serta GitOps.
- Grafana, Mimir, Loki, Tempo, dan Alloy untuk observability.
- OpenTofu dan Ansible untuk IaC.
- Trivy, Cilium, KEDA, Velero, dan tool lain sesuai minggu yang dipelajari.

Pastikan cluster aktif sebelum lab:

```bash
kubectl get nodes
kubectl get namespaces
```

Gunakan cluster lokal atau environment khusus latihan. Jangan menjalankan
operasi destruktif, eksperimen chaos, penghapusan resource, atau pengujian
credential pada production.

### 3. Jalankan manifest dengan terkontrol

Baca manifest sebelum menjalankannya, terutama namespace, image, resource
request/limit, volume, Service, Ingress, dan konfigurasi Secret.

Contoh pola praktik:

```bash
# Tinjau perubahan tanpa menerapkannya
kubectl diff -f minggu-01/manifests/

# Terapkan manifest
kubectl apply -f minggu-01/manifests/

# Amati resource
kubectl get all -n mini-prod
kubectl get events -n mini-prod --sort-by=.lastTimestamp

# Bersihkan resource setelah lab selesai
kubectl delete -f minggu-01/manifests/
```

Sesuaikan namespace dan nama resource dengan isi manifest. Untuk setiap lab,
ikuti perintah cleanup yang ada di modul agar environment tidak menyimpan
resource eksperimen.

### 4. Gunakan pola belajar teori lalu praktik

Untuk setiap modul, gunakan siklus berikut:

```text
Baca konsep -> prediksi perilaku -> jalankan lab -> amati output
-> dokumentasikan temuan -> lakukan cleanup -> ulangi dengan variasi
```

Saat menjalankan lab insiden, gunakan kerangka:

```text
Symptoms -> Investigation -> Root Cause -> Mitigation -> Prevention
```

Hindari langsung mengubah konfigurasi sampai gejala dan bukti tercatat. Tujuan
latihan adalah membangun kemampuan diagnosis, bukan hanya membuat status Pod
kembali menjadi `Running`.

### 5. Simpan artefak pembelajaran

Simpan hasil berikut di branch atau direktori catatan pribadi:

- Manifest atau konfigurasi yang Anda ubah.
- Output penting dari `kubectl`, Helm, OpenTofu, Ansible, atau tool lain.
- Screenshot atau export dashboard saat observasi.
- Timeline insiden dan hipotesis yang diuji.
- Runbook, postmortem, RTO/RPO, SLO, dan action item.
- Catatan keputusan: mengapa tool, threshold, resource limit, atau policy dipilih.

Artefak ini berguna sebagai portofolio teknis dan bukti bahwa lab benar-benar
dipahami.

## Jalur Belajar yang Disarankan

### Jalur pemula

Mulai dari [Tahap 0](./tahap-0/README.md), terutama modul Linux, Networking, dan Git.
Setelah acceptance gate terpenuhi, lanjutkan Minggu 1-4 dengan fokus pada
Kubernetes, deployment aplikasi, Helm, dan GitOps. Jangan memasang seluruh
observability stack sebelum workload dasar dapat di-deploy dan di-debug secara
manual.

### Jalur observability dan operasi

Lanjutkan Minggu 5-12. Bangun tiga pilar observability secara berurutan:
metrics, logs, lalu traces. Setelah itu gunakan semuanya untuk alerting,
troubleshooting, SLO, dan production simulation.

### Jalur platform production-ready

Lanjutkan Minggu 13-18 setelah capstone pertama selesai. Terapkan HA, IaC,
security, networking, autoscaling, lalu backup dan DR. Setiap komponen baru
harus memiliki alasan adopsi, owner, rollback, dan runbook.

### Jalur roadmap lanjutan

Minggu 19-24 digunakan untuk progressive delivery, supply-chain security, cost
and capacity, chaos engineering, governance, upgrade, dan final game day.
Gunakan [`README-Lanjutan.md`](./README-Lanjutan.md) sebagai acuan scope dan
acceptance gate sebelum menambahkan implementasi modul baru.

## Aturan Praktik dan Keselamatan

- Kunci versi image, chart, dan dependency bila lab akan diulang.
- Jangan menyimpan password, token, private key, atau credential asli di Git.
- Perlakukan Secret Kubernetes sebagai data sensitif walaupun nilainya Base64.
- Gunakan namespace khusus lab dan label resource dengan jelas.
- Verifikasi dampak `kubectl delete`, `drain`, `tofu apply`, dan operasi restore.
- Uji backup dengan restore; backup yang belum pernah direstore belum terbukti.
- Tulis perubahan infrastructure dan policy dalam Git agar dapat diaudit.
- Bersihkan resource setelah eksperimen untuk menghindari konflik dan konsumsi resource.

## Indikator Lulus Keseluruhan

Kurikulum dianggap selesai jika Anda dapat menunjukkan bahwa platform:

- Dapat dibuat dan dideploy ulang dari kode.
- Memiliki delivery yang terversi dan dapat di-rollback.
- Memiliki metrics, logs, traces, dashboard, dan alert yang dapat ditindaklanjuti.
- Memiliki SLO, runbook, error budget, dan postmortem.
- Tetap melayani workload saat node atau Pod mengalami gangguan terkontrol.
- Memiliki policy keamanan, scan image, dan akses least privilege.
- Dapat melakukan autoscaling berdasarkan beban atau event.
- Memiliki backup, restore evidence, RPO/RTO, dan DR drill.
- Dapat melewati final game day tanpa mengandalkan langkah manual yang tidak terdokumentasi.

## Referensi

- [Roadmap SRE lanjutan](./README-Lanjutan.md)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Google SRE Book](https://sre.google/sre-book/)
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Grafana Documentation](https://grafana.com/docs/)
