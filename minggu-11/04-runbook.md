# Modul 04 — Runbook: Prosedur Standar untuk Tim On-Call

> **Satu kalimat:** **Runbook** adalah dokumen **cheat sheet** yang bisa dijalankan (executable) oleh engineer on-call PUKUL 3 PAGI untuk merespons insiden tanpa harus berpikir panjang — berisi langkah diagnosis, mitigasi, dan eskalasi.

Salah satu perbedaan utama antara tim DevOps pemula dan tim SRE yang mature adalah **dokumentasi runbook**. Tanpa runbook, setiap insiden adalah "reinvent the wheel" — on-call harus men-debug dari nol, panik, dan membuat keputusan buruk. **Dengan runbook, on-call jadi seperti pilot yang punya checklist.**

---

## 🎯 Learning Outcomes

1. Memahami **anatomi runbook** yang efektif (bukan sekadar "troubleshooting guide").
2. Mendesain runbook dengan **decision tree** dan **ekspektasi output** di setiap langkah.
3. Membuat **runbook untuk 5 insiden umum** Kubernetes (Pod Crash, OOM, DNS, PVC, Deployment Stuck).
4. Mengintegrasikan runbook ke Alertmanager (runbook URL di alert annotation).
5. Menulis **escalation policy** yang jelas dan **postmortem template**.

---

## 1. 🧠 Filosofi: Runbook Bukan "Tutorial", Tapi "Checklist Pilot"

Ada perbedaan halus tapi penting:

```text
┌──────────────────────────────────────────────────────────────────┐
│ Tutorial (BUKAN runbook)         │ Runbook (Yang baik)           │
├───────────────────────────────────┼───────────────────────────────┤
│ "Pelajari cara kerja Kubernetes" │ "Pod X error → jalankan Y"   │
│ Panjang, konseptual              │ Singkat, eksekusi langsung    │
│ Untuk belajar                   │ Untuk on-call 3 AM            │
│ Mengajarkan WHY                  │ Hanya WHAT dan HOW            │
│ Tidak ada deadline               │ Ada target SLA resolution     │
└───────────────────────────────────┴───────────────────────────────┘
```

**Analogi:**
- **Tutorial** = buku teori pilot tentang aerodinamika.
- **Runbook** = checklist pre-takeoff yang ada di kokpit.

---

## 2. 🏗️ Anatomi Runbook yang Efektif

Setiap runbook yang baik memiliki **7 bagian standar**:

```text
┌──────────────────────────────────────────────────────────────────┐
│ RUNBOOK STRUCTURE                                                │
│                                                                  │
│  1. TITLE & METADATA         (Apa, severity, owner)              │
│  2. SYMPTOMS                 (Bagaimana user melihat ini)        │
│  3. DIAGNOSIS STEPS          (Cara cek apakah insiden ini)       │
│  4. MITIGATION STEPS         (Cara pulihkan servis, BUKAN fix)   │
│  5. ESCALATION               (Kapan harus kontak siapa)          │
│  6. PREVENTION               (Cara cegah terulang — link PR)     │
│  7. RELATED LINKS            (Dashboard, alert, log query)       │
└──────────────────────────────────────────────────────────────────┘
```

### 2.1 Template Standar (Markdown)

```markdown
# RB-001: Pod CrashLoopBackOff Recovery

| Metadata | Value |
|---|---|
| **Severity** | P1 (Customer-facing) |
| **Owner** | SRE Team (@sre-oncall) |
| **First Response SLA** | 5 menit |
| **Resolution SLA** | 30 menit |
| **Last Tested** | 2026-08-01 |
| **Last Updated** | 2026-08-10 |

## Symptoms
- Alert: `KubePodCrashLooping` firing
- Customer report: HTTP 503
- Dashboard: <link>

## Diagnosis
[Step-by-step command + expected output]

## Mitigation
[Quick fix to restore service]

## Escalation
- 0-15 min: Try mitigation
- 15-30 min: Ping @sre-lead (Slack)
- 30+ min: Page VP Engineering
```

---

## 3. 📋 5 Runbook Wajib untuk Mini Production Platform

### RB-001: Pod CrashLoopBackOff

| Metadata | Value |
|---|---|
| **Alert** | `KubePodCrashLooping` |
| **Severity** | P1 (jika user-facing) / P2 (jika internal) |
| **First Response** | 5 menit |

#### Symptoms
- Alert firing: `KubePodCrashLooping{namespace="default"}`
- HTTP 503 dari ingress
- Replica count tidak sesuai desired state

#### Diagnosis (30-60 detik)
```bash
# Step 1: Lihat status semua Pod di namespace terkait
$ kubectl get pods -n <NAMESPACE>
NAME                          READY   STATUS             RESTARTS   AGE
checkout-service-xxx          0/1     CrashLoopBackOff   8          5m

# EXPECTED: Ada Pod dengan status CrashLoopBackOff dan RESTARTS > 3

# Step 2: Inspect event untuk lihat exit code
$ kubectl describe pod <POD_NAME> -n <NAMESPACE> | grep -A 3 "Last State"
Last State:     Terminated
  Reason:       Error
  Exit Code:    1

# EXPECTED: Exit code 1 (app error) atau 137 (OOM) atau 139 (segfault)

# Step 3: Lihat log container
$ kubectl logs <POD_NAME> -n <NAMESPACE> --previous --tail=50

# EXPECTED: Stack trace, panic, atau error message jelas
```

#### Mitigation (Pilih Berdasarkan Exit Code)

| Exit Code | Meaning | Action |
|---|---|---|
| **0** | Normal exit (app selesai) | Cek konfigurasi `restartPolicy` |
| **1** | Application error | Perbaiki code, rollback deployment |
| **137** | OOMKilled (SIGKILL) | Naikkan memory limit |
| **139** | Segfault (SIGSEGV) | Cek native library / restart |
| **143** | SIGTERM (graceful shutdown) | Biasanya karena rolling update |

**Mitigation cepat untuk restore service:**
```bash
# Option 1: Rollback ke versi sebelumnya (jika deploy baru)
$ kubectl rollout undo deployment/checkout-service -n <NAMESPACE>
deployment.apps/checkout-service rolled back

# Option 2: Naikkan resource limit (jika OOM)
$ kubectl set resources deployment/checkout-service \
    -n <NAMESPACE> \
    --limits=memory=512Mi,cpu=500m
```

#### Escalation
- 0-15 min: Coba `rollout undo`
- 15-30 min: Slack `#sre-oncall` dengan @channel
- 30+ min: Page SRE Manager (PagerDuty escalation)

#### Prevention
- Tambah readiness probe (lihat RB-005)
- Pasang pre-commit hook untuk cek exit code 1
- Implementasi canary deployment

---

### RB-002: Node Disk Pressure

| Metadata | Value |
|---|---|
| **Alert** | `KubeletDiskPressure` |
| **Severity** | P2 |
| **First Response** | 10 menit |

#### Symptoms
- Node status `Ready,SchedulingDisabled`
- New Pod stuck `Pending`
- Alert: `node_filesystem_avail_bytes < 10%`

#### Diagnosis
```bash
# Step 1: Identifikasi node yang penuh
$ kubectl get nodes
NAME           STATUS                     ROLES                  AGE
laptop-node-1  Ready,SchedulingDisabled   control-plane,master   30d

# EXPECTED: Ada node dengan SchedulingDisabled

# Step 2: Cek kondisi node
$ kubectl describe node laptop-node-1 | grep -A 8 Conditions
Conditions:
  Type                 Status  Reason          Message
  ----                 ------  ------          -------
  Ready                False   DiskPressure    kubelet has disk pressure

# EXPECTED: DiskPressure = True

# Step 3: Cari direktori terbesar
$ ssh laptop-node-1
$ sudo du -sh /var/lib/containerd /var/log /tmp | sort -hr
30G   /var/lib/containerd   ← culprit biasanya image cache
8.5G  /var/log
1.2G  /tmp
```

#### Mitigation
```bash
# Step 1: Cordon node (stop receiving new Pod)
$ kubectl cordon laptop-node-1

# Step 2: Prune image cache
$ sudo crictl rmi --prune

# Step 3: Truncate log file yang membengkak (JANGAN rm!)
$ sudo truncate -s 0 /var/log/containers/<app>-*.log

# Step 4: Restart kubelet (last resort)
$ sudo systemctl restart k3s

# Step 5: Verify disk free
$ df -h /
# EXPECTED: Use% < 85%

# Step 6: Uncordon node
$ kubectl uncordon laptop-node-1
```

#### Prevention
- Pasang alert pada 80% (bukan 95%)
- Konfigurasi `image-gc-high-threshold=80` di kubelet
- Implementasi logrotate DaemonSet
- Capacity planning: provision 2x storage yang terpakai

---

### RB-003: CoreDNS Down

| Metadata | Value |
|---|---|
| **Alert** | `CoreDNSDown` atau `KubeDNSDown` |
| **Severity** | P1 (cluster-wide impact) |
| **First Response** | 2 menit |

#### Symptoms
- Semua Pod gagal resolve DNS: `no such host`
- Aplikasi yang sebelumnya normal tiba-tiba error
- Alert: `CoreDNSDown` atau `coredns_*` metrics = 0

#### Diagnosis
```bash
# Step 1: Cek status CoreDNS
$ kubectl get pods -n kube-system -l k8s-app=kube-dns
NAME                       READY   STATUS             RESTARTS   AGE
coredns-6743829-x8291      0/1     CrashLoopBackOff   12         1h
coredns-6743829-p210a      1/1     Running            0          1h

# EXPECTED: Minimal 1 replica CrashLoopBackOff

# Step 2: Lihat log
$ kubectl logs -n kube-system -l k8s-app=kube-dns --tail=50

# EXPECTED: Syntax error Corefile atau upstream timeout
```

#### Mitigation
```bash
# Step 1: Scale up CoreDNS untuk absorb load sementara
$ kubectl scale deployment coredns -n kube-system --replicas=4

# Step 2: Restart semua Pod CoreDNS untuk recover dari stale state
$ kubectl rollout restart deployment coredns -n kube-system

# Step 3: Verify DNS resolution dari debug pod
$ kubectl run dnstest --image=registry.k8s.io/e2e-test-images/jessie-dnsutils:1.3 \
    --rm -it --restart=Never -- nslookup kubernetes.default
Server:    10.96.0.10
Address 1: 10.96.0.10 kube-dns.svc.cluster.local
Name:      kubernetes.default
Address 1: 10.96.0.1
# EXPECTED: Tidak ada timeout
```

#### Prevention
- Pasang PodDisruptionBudget untuk CoreDNS (min 2 always available)
- Backup Corefile ke Git (GitOps)
- Alerting pada restart count

---

### RB-004: PVC Stuck Pending

| Metadata | Value |
|---|---|
| **Alert** | `KubePersistentVolumeClaimPending` |
| **Severity** | P2 (database) / P3 (cache) |
| **First Response** | 10 menit |

#### Symptoms
- Pod yang pakai PVC stuck `ContainerCreating`
- Event: `FailedMount` atau `FailedScheduling`
- Alert: PVC pending > 15 menit

#### Diagnosis
```bash
# Step 1: Cek status PVC
$ kubectl get pvc -A
NAME             STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
postgres-data    Pending   <none>                          local-path     30m

# EXPECTED: PVC dengan STATUS=Pending dan VOLUME=<none>

# Step 2: Inspect event
$ kubectl describe pvc postgres-data -n database
Events:
  Type     Reason              Age   From                         Message
  ----     ------              ----  ----                         -------
  Warning  ProvisioningFailed  30m   persistentvolume-controller  storageclass "fast-ssd" not found

# EXPECTED: Pesan error yang jelas
```

#### Mitigation (Berdasarkan Root Cause)

| Root Cause | Fix |
|---|---|
| **StorageClass not found** | Ganti SC reference atau create SC |
| **Insufficient capacity** | Expand node disk atau kurangi PVC size |
| **VolumeBindingMode WaitForFirstConsumer** | Deploy Pod yang reference PVC untuk trigger binding |
| **Access mode mismatch** | Ganti PVC accessModes ke RWO/RWX yang sesuai |
| **No available PersistentVolume** | Cek apakah ada PV unbound |

**Quick fix jika PVC urgent:**
```bash
# Option 1: Delete dan recreate dengan StorageClass yang ada
$ kubectl delete pvc postgres-data -n database
$ kubectl apply -f pvc-fixed.yaml  # dengan storageClassName: local-path

# Option 2: Bind manual ke PV existing
$ kubectl edit pvc postgres-data -n database
# Tambahkan spec.volumeName: <existing-pv-name>
```

---

### RB-005: Deployment Stuck Rolling Update

| Metadata | Value |
|---|---|
| **Alert** | `KubeDeploymentReplicasMismatch` |
| **Severity** | P2 |
| **First Response** | 10 menit |

#### Symptoms
- `kubectl rollout status deployment/X` stuck
- New ReplicaSet ada, tapi Replicas tidak naik
- Alert: deployment replicas != desired

#### Diagnosis
```bash
# Step 1: Cek status rollout
$ kubectl rollout status deployment/checkout-service
Waiting for deployment "checkout-service" rollout to finish: 2 out of 3 new replicas updated...

# Step 2: Cek status ReplicaSet
$ kubectl get rs -l app=checkout-service
NAME                          DESIRED   CURRENT   READY   AGE
checkout-service-7d8c9b5f8    3         2         2       10m   ← new
checkout-service-6b7c4d5e9    0         0         0       30m   ← old (terminated)

# Step 3: Cek Pod baru
$ kubectl get pods -l app=checkout-service
NAME                              READY   STATUS    RESTARTS   AGE
checkout-service-7d8c9b5f8-xxx    1/1     Running   0          10m
checkout-service-7d8c9b5f8-yyy    1/1     Running   0          10m
checkout-service-7d8c9b5f8-zzz    0/1     Pending   0          10m   ← ini yang stuck

# Step 4: Describe Pod pending
$ kubectl describe pod checkout-service-7d8c9b5f8-zzz
Events:
  Type     Reason            Age   From               Message
  ----     ------            ----  ----               -------
  Warning  FailedScheduling  10m   default-scheduler  0/1 nodes are available: 1 Insufficient cpu.

# EXPECTED: Resource tidak cukup, atau PersistentVolumeClaim unbind, atau taints
```

#### Mitigation
```bash
# Option 1: Rollback ke versi sebelumnya
$ kubectl rollout undo deployment/checkout-service

# Option 2: Pause rollout, fix issue, resume
$ kubectl rollout pause deployment/checkout-service
# ... fix image / resources / volume ...
$ kubectl rollout resume deployment/checkout-service

# Option 3: Force rollout baru
$ kubectl rollout restart deployment/checkout-service
```

---

## 4. 🔗 Integrasi Runbook ke Alertmanager

Tambahkan runbook URL di setiap Prometheus alert annotation:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: alerts-with-runbook
spec:
  groups:
  - name: kubernetes_alerts
    rules:
    - alert: KubePodCrashLooping
      expr: |
        rate(kube_pod_container_status_restarts_total[10m]) > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Pod {{ $labels.pod }} crash looping"
        runbook_url: "https://wiki.internal/runbooks/rb-001-pod-crash"
        dashboard_url: "https://grafana.internal/d/cluster-overview"
        logs_url: "https://grafana.internal/explore?ds=loki&query={{ $labels.pod }}"
```

Alertmanager akan otomatis menambahkan **link runbook** di Slack/PagerDuty notification:

```text
🚨 CRITICAL: Pod checkout-service-xxx crash looping
   Summary: Pod in namespace default restarting 8 times in 10m
   📖 Runbook: https://wiki.internal/runbooks/rb-001-pod-crash
   📊 Dashboard: https://grafana.internal/d/cluster-overview
   📝 Logs: https://grafana.internal/explore?...
```

---

## 5. 🛡️ Escalation Policy

### 5.1 Tier-Based Escalation

```text
┌──────────────────────────────────────────────────────────────────┐
│ ESCALATION MATRIX                                                │
│                                                                  │
│  P0 INCIDENT (Revenue-impacting)                                 │
│  ├── 0-5 min    : SRE on-call responds                           │
│  ├── 5-15 min   : SRE Lead + Product Manager joined              │
│  ├── 15-30 min  : VP Engineering paged                           │
│  └── 30+ min    : CEO notification (via Slack)                   │
│                                                                  │
│  P1 INCIDENT (Service degraded)                                  │
│  ├── 0-15 min   : SRE on-call responds                           │
│  ├── 15-30 min  : SRE Lead notified                              │
│  └── 30+ min    : VP Engineering paged                           │
│                                                                  │
│  P2 INCIDENT (Internal tooling)                                  │
│  ├── 0-1 hour   : SRE on-call responds (during business hours)   │
│  └── 1+ hour    : Reschedule to next business day                │
└──────────────────────────────────────────────────────────────────┘
```

### 5.2 On-Call Schedule

```yaml
# Contoh rotation 1 minggu (7 orang)
rotation:
  - week: 2026-08-04
    primary: "@alice"
    secondary: "@bob"
  - week: 2026-08-11
    primary: "@charlie"
    secondary: "@diana"
  - ...
```

---

## 6. 📦 Format Penyimpanan Runbook

### 6.1 Opsi Storage

| Opsi | Pro | Kontra |
|---|---|---|
| **Git repository** (`runbooks/`) | Version controlled, PR review | Tidak ada UI search yang bagus |
| **Confluence/Notion** | Search bagus, collaborative | Bisa out-of-sync dengan actual command |
| **Grafana annotations** | Inline dengan dashboard | Hanya untuk konteks dashboard |
| **Dedicated tool (Slab, Tettra)** | Search + analytics | Vendor lock-in, biaya |

**Rekomendasi:** Gunakan **Git repository** untuk runbook yang executable (kode YAML, command), dan **Confluence** untuk runbook konseptual.

### 6.2 Repository Structure

```text
runbooks/
├── README.md                  # Index semua runbook
├── RB-001-pod-crash.md        # Bisa dijalankan langsung
├── RB-002-disk-pressure.md
├── RB-003-coredns-down.md
├── RB-004-pvc-pending.md
├── RB-005-deployment-stuck.md
├── templates/
│   └── runbook-template.md    # Template standar
└── escalation/
    └── oncall-rotation.yaml   # Schedule rotasi
```

---

## 7. 🧪 Latihan: Buat Runbook Pertama Anda

**Tugas:** Buat runbook untuk insiden yang Anda tangani di Week 9 atau Week 10. Format:

```markdown
# RB-XXX: [Nama Insiden]

| Metadata | Value |
|---|---|
| Alert | <alert_name> |
| Severity | <P0/P1/P2> |
| First Response | <menit> |

## Symptoms
- ...

## Diagnosis
```bash
# Setiap command + expected output
```

## Mitigation
```bash
# Command yang bisa langsung dijalankan
```

## Escalation
- ...
```

Upload ke `minggu-11/runbooks/` di repo Anda.

---

## 8. 📋 Cheat Sheet: Anatomy of a Good Runbook

```text
GOOD RUNBOOK                          | BAD RUNBOOK
--------------------------------------|----------------------------------------
✓ Exit code 1 → fix env vars          | ❌ "Fix the configuration"
✓ kubectl rollout undo deploy/X       | ❌ "Rollback when broken"
✓ df -h / shows Use% 95%              | ❌ "Check disk"
✓ Expected: HTTP 200                  | ❌ "Service should be back"
✓ Link to dashboard                   | ❌ "Look at metrics"
✓ Tested 1 bulan lalu                 | ❌ Never tested (probably wrong)
✓ Owner: @sre-lead                    | ❌ No owner
```

---

## 9. ✏️ Latihan Mandiri

1. **Tulis runbook** untuk insiden `N+1 Query` dari Week 10 — bagaimana cara cepat on-call melakukan kill query?
2. **Buat escalation matrix** untuk tim 2-engineer (siapa yang dipanggil kapan).
3. **Test runbook Anda** di staging environment, ukur berapa lama eksekusi.
4. **Pasang runbook_url annotation** di semua alert Kubernetes Anda.

---

**Lanjut ke Modul 05:** [05-postmortem.md](./05-postmortem.md) — cara menulis postmortem yang blameless dan menghasilkan action items yang actionable.
