# Roadmap SRE Lanjutan - 12 Minggu Berikutnya

Roadmap ini melanjutkan Minggu 01 sampai Minggu 12. Fokusnya adalah capability production yang belum cukup kuat pada mini platform dasar:


## Tujuan tahap lanjutan

Pada akhir minggu 24, target platform bukan sekadar "bisa deploy", tetapi mampu:

- Dibuat ulang dari kode.
- Dijalankan pada lebih dari satu Node.
- Diamankan dengan policy dan supply-chain verification.
- Di-scale berdasarkan beban.
- Di-backup dan di-restore.
- Di-release secara canary/blue-green.
- Diuji dengan failure dan chaos yang terkontrol.
- Diukur cost, capacity, SLO, dan error budget.
- Dioperasikan melalui runbook dan game day.

## Cara menggunakan roadmap

1. Kerjakan satu minggu sampai deliverable dan acceptance gate terpenuhi.
2. Jangan memasang semua tool sekaligus pada laptop.
3. Simpan konfigurasi di Git dan gunakan ArgoCD setelah bootstrap.
4. Setiap tool baru harus memiliki alasan, owner, rollback, dan runbook.
5. Uji destructive operation hanya pada environment lab.

## Ringkasan 12 minggu

| Minggu | Fokus | Kemampuan utama | Artefak utama |
| --- | --- | --- | --- |
| 1 | Fundamental Container & Kubernetes | Container, Pod, Deployment, Service, Ingress | Manifest Kubernetes dasar |
| 2 | Kubernetes Workload | ConfigMap, Secret, PVC, probes, workload controller | Go app dan manifest production-oriented |
| 3 | Helm | Chart, template, values, release, rollback | Helm chart Go app |
| 4 | GitLab + GitOps | CI, Runner, registry, ArgoCD, self-heal | GitLab CI dan ArgoCD Application |
| 5 | Metrics | Alloy, Mimir, PromQL, kube-state-metrics | Metrics pipeline dan dashboard cluster |
| 6 | Logging | JSON log, Loki, LogQL, label policy | Alloy log pipeline dan logging dashboard |
| 7 | Tracing | OpenTelemetry, OTLP, Tempo, TraceQL | Trace HTTP/database/Redis |
| 8 | Dashboard & Alert | RED, USE, Golden Signals, routing | Dashboard service dan alert rules |
| 9 | Incident Simulation I | CrashLoop, OOM, Pending, image pull, mount | Incident runbook dasar |
| 10 | Incident Simulation II | CPU, memory, disk, DNS, network, latency, deadlock | Multi-signal investigation |
| 11 | Reliability Engineering | SLI, SLO, SLA, error budget, k6 | SLO report, load test, recording rules |
| 12 | Production Simulation | Delivery sampai incident recovery | Game day dan postmortem |
| 13 | HA Cluster | Multi-node k3s, etcd quorum, PDB | HA cluster dan failure evidence |
| 14 | Infrastructure as Code | OpenTofu, Ansible, state, drift | IaC starter dan generated inventory |
| 15 | Security Baseline | RBAC, PSA, Kyverno, cert-manager, scanning | Security policies dan secure workload |
| 16 | Networking | Cilium, Hubble, DNS, NetworkPolicy, egress | Traffic matrix dan network policies |
| 17 | Autoscaling | HPA, VPA, KEDA, stabilization | Scaling manifests dan scaling curve |
| 18 | Backup & DR | Velero, object storage, RPO, RTO, restore | Backup schedule dan restore evidence |
| 19 | Progressive Delivery | Argo Rollouts, canary, blue-green | Rollout dan AnalysisTemplate |
| 20 | Supply Chain | Trivy, Syft, Grype, Cosign, Renovate | SBOM, signature, CI security stage |
| 21 | Cost & Capacity | OpenCost, rightsizing, forecast, budget | Cost query pack dan capacity model |
| 22 | Chaos Engineering | Chaos Mesh, hypothesis, blast radius | Chaos experiment dan report |
| 23 | Governance & Upgrade | AppProject, ownership, version pinning, upgrade | Governance policy dan compatibility matrix |
| 24 | Final Game Day | Semua capability | Final assessment dan postmortem |



## Prinsip pemilihan tool

- Pilih tool yang menyelesaikan masalah nyata, bukan yang paling banyak di-install.
- Utamakan proyek CNCF atau komunitas upstream yang aktif.
- Cek release cadence, dokumentasi, security advisories, dan maintainer.
- Bedakan open source, source-available, dan managed service.
- Kunci versi chart dan image.
- Jangan memakai lisensi sebagai satu-satunya indikator kualitas; cek juga operational fit.

## Catatan lisensi

Terraform sangat banyak dipakai, tetapi Terraform modern menggunakan lisensi BUSL, bukan lisensi open source OSI. Untuk roadmap yang mengutamakan open source, gunakan OpenTofu sebagai default dan pelajari Terraform untuk kompatibilitas pasar kerja. Selalu cek lisensi dan kebijakan organisasi sebelum adopsi.
