# Minggu 9 — Modul 06: Incident FailedMount

> **"Pod Pending lama, akhirnya Running tapi container crash. Log bilang tidak bisa baca file."**

FailedMount artinya K8s **tidak bisa mount volume** yang dideklarasikan di spec pod. Ini sering terjadi di stateful workload (database, cache) yang butuh persistent storage.

Penyebab umum:

1. **PVC tidak ditemukan** — nama PVC salah ketik atau PVC belum dibuat
2. **PV tidak tersedia** — tidak ada PV yang match dengan PVC request
3. **StorageClass tidak ada** — PVC reference storageClass yang tidak ter-install
4. **Permission issue** — NFS/Ceph share tidak writable oleh user container
5. **Node tidak bisa akses storage backend** — misal NFS server down atau network issue

Analogi: seperti kamu sewa rumah kos, dikasih kunci tapi pintunya tidak ada (PVC not found), atau kunci tidak cocok (claimName mismatch), atau petaknya sedang direnovasi (PV unavailable).

---

## 🎯 Tujuan Modul

1. Membedakan 4 jenis FailedMount berdasarkan event message
2. Memahami **PVC lifecycle**: Pending → Bound, dan kenapa bisa stuck
3. Membuat **PVC + StorageClass** yang benar
4. Memilih **mitigation** (buat PVC, hapus volume ref, fix StorageClass)
5. Menerapkan **backup strategy** (Velero) dan **naming convention** untuk prevention

---

## 📦 Simulasi — FailedMount Karena PVC Tidak Ada

### File: `manifests/05-failedmount-missing-pvc.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: failedmount-app
  namespace: insiden-lab
spec:
  replicas: 1
  selector:
    matchLabels:
      app: failedmount-app
  template:
    metadata:
      labels:
        app: failedmount-app
    spec:
      containers:
      - name: app
        image: nginx:1.27-alpine
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: data-pvc-tidak-ada    # ← PVC ini tidak pernah dibuat
```

### Deploy & Observe

```bash
kubectl apply -f manifests/05-failedmount-missing-pvc.yaml

kubectl get pods -n insiden-lab -l app=failedmount-app -w
```

**Output:**
```
NAME                              READY   STATUS              RESTARTS   AGE
failedmount-app-7f9c8d9f8-abcd1   0/1     ContainerCreating   0          5s
failedmount-app-7f9c8d9f8-abcd1   0/1     ContainerCreating   0          30s
```

**Catatan:** Stuck di `ContainerCreating` lama, tidak pernah jadi Running.

---

## 🔍 Symptoms

### Symptom 1 — Stuck di `ContainerCreating`

```bash
$ kubectl get pods -n insiden-lab -l app=failedmount-app
NAME                              READY   STATUS              RESTARTS   AGE
failedmount-app-7f9c8d9f8-abcd1   0/1     ContainerCreating   0          5m    ← stuck 5 menit
```

**Bedanya dengan CrashLoopBackOff:** Status ContainerCreating (bukan CrashLoopBackOff). Container **belum bisa jalan** karena mount gagal.

### Symptom 2 — Alert firing (kamu setup di Minggu 8)

```
[FIRING] PodsFailedMount
  namespace = insiden-lab
  pod       = failedmount-app-xxx
  pvc       = data-pvc-tidak-ada
  reason    = FailedMount
  severity  = warning
```

### Symptom 3 — Log kosong atau error permission

Karena container tidak jalan, log mungkin kosong. Kalau mount berhasil tapi data kosong/permission ditolak:
```
nginx: [emerg] open() "/data/index.html" failed (2: No such file or directory)
```

---

## 🕵️ Investigation

### Step 1 — describe pod

```bash
$ kubectl describe pod failedmount-app-7f9c8d9f8-abcd1 -n insiden-lab
```

**Output penting:**

```
Containers:
  app:
    ...
    State:          Waiting
      Reason:       ContainerCreating
    Ready:          False

Events:
  Type     Reason              Age    From               Message
  ----     ------              ----   ----               -------
  Normal   Scheduled           30s    default-scheduler  Successfully assigned ...
  Warning  FailedMount         25s    kubelet            MountVolume.SetUp failed for volume
                                                       "pvc-xxxxx" :
                                                       persistentvolumeclaims "data-pvc-tidak-ada"
                                                       not found
  Warning  FailedMount         20s    kubelet            Unable to attach or mount volumes:
                                                       unmounted volumes=[data]
                                                       failed to process volume mounts for volume
                                                       "pvc-xxxxx" : pv not found
```

**Pesan kunci:**
- `persistentvolumeclaims "data-pvc-tidak-ada" not found` → PVC dengan nama itu tidak ada
- `pv not found` → tidak ada PV yang match

### Step 2 — Cek apakah PVC ada

```bash
$ kubectl get pvc -n insiden-lab
# No resources found in insiden-lab namespace.

$ kubectl get pvc -A
# Cek apakah ada PVC serupa di namespace lain (mungkin salah namespace)
```

### Step 3 — Cek status PVC (kalau PVC ada tapi Pending)

```bash
$ kubectl get pvc data-pvc -n insiden-lab
# NAME       STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
# data-pvc   Pending                                      slow           5m

$ kubectl describe pvc data-pvc -n insiden-lab
```

**Output (PVC Pending):**
```
Events:
  Type     Reason              Age   From                         Message
  ----     ------              ----  ----                         -------
  Warning  ProvisioningFailed  1m    persistentvolume-controller  storageclass "slow" not found
```

### Step 4 — Cek StorageClass

```bash
$ kubectl get storageclass
# NAME                   PROVISIONER                RECLAIMPOLICY   VOLUMEBINDINGMODE
# local-path (default)   rancher.io/local-path      Delete          WaitForFirstConsumer
```

**Catatan:** `local-path` StorageClass sudah terinstall otomatis di k3s. Tapi kalau PVC reference StorageClass yang tidak ada, akan stuck Pending.

### Step 5 — Lihat events lengkap

```bash
$ kubectl get events -n insiden-lab --field-selector reason=FailedMount --sort-by=.lastTimestamp
```

---

## 🎯 Root Cause Analysis

### 5 Whys untuk Missing PVC

```
Problem: Pod failedmount-app ContainerCreating 5 menit

  Why 1: Kenapa stuck ContainerCreating?
    → Kubelet gagal mount volume "data"
  
  Why 2: Kenapa gagal mount?
    → PVC "data-pvc-tidak-ada" not found
  
  Why 3: Kenapa PVC tidak ada?
    → PVC-nya memang tidak pernah dibuat
  
  Why 4: Kenapa deployment refer ke PVC yang tidak ada?
    → Manifest deployment dan manifest PVC dibuat terpisah,
       tidak ada validasi referensi
  
  Why 5: Kenapa tidak ada validasi?
    → Tidak ada GitOps check atau helm lint untuk validasi reference

🎯 ROOT CAUSE: PVC reference broken — tidak dibuat
```

### Tabel Diagnosis FailedMount

| Event Message Fragment | Root Cause | Solusi |
|---|---|---|
| `persistentvolumeclaims "X" not found` | PVC name salah atau missing | Buat PVC atau fix nama |
| `storageclass "X" not found` | StorageClass tidak terinstall | Install provisioner atau pakai SC yang ada |
| `Failed to provision volume with StorageClass "X"` | Provisioner error | Cek log provisioner |
| `mount.nfs: Connection refused` | NFS server down | Cek NFS server |
| `mount.nfs: access denied` | NFS export permission | Fix /etc/exports |
| `Volume has not been provisioned yet` | PV provisioning masih proses | Tunggu atau cek provisioner |
| `node X cannot access storage backend` | Node-specific storage issue | Cek CSI driver di node |

---

## 🛠️ Mitigation

### Opsi A — Buat PVC yang Hilang (Paling Umum)

```bash
# Buat PVC sederhana dengan local-path StorageClass (default di k3s)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc-tidak-ada    # ← SAMA dengan yang di-reference
  namespace: insiden-lab
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  # storageClassName: local-path   # opsional, default sudah cukup
EOF

# Tunggu beberapa detik, kubelet akan retry mount
kubectl get pods -n insiden-lab -l app=failedmount-app -w
```

**Output yang diharapkan:**
```
NAME                              READY   STATUS    RESTARTS   AGE
failedmount-app-7f9c8d9f8-abcd1   1/1     Running   0          30s    ← akhirnya Running!
```

### Opsi B — Fix StorageClass Name yang Salah

```bash
# Lihat StorageClass yang tersedia
kubectl get storageclass
# Pilih yang ada, misal "local-path"

# Patch PVC untuk pakai SC yang valid
kubectl patch pvc data-pvc -n insiden-lab --type=json   -p='[{"op":"replace","path":"/spec/storageClassName","value":"local-path"}]'
```

### Opsi C — Hapus Volume Reference (kalau app tidak butuh PVC)

```bash
# Patch deployment untuk hapus volume reference
kubectl patch deployment/failedmount-app -n insiden-lab --type=json   -p='[{"op":"remove","path":"/spec/template/spec/volumes"}]'
```

### Opsi D — Buat PV Manual (untuk SC-less cluster)

```bash
# Kadang di cluster tanpa dynamic provisioning
# Harus buat PV manual
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: manual-pv-1
spec:
  capacity:
    storage: 1Gi
  accessModes:
  - ReadWriteOnce
  hostPath:
    path: /tmp/pv-data
EOF

# PVC dengan storageClassName: "" akan bind ke PV manual
```

### Opsi E — Recovery NFS/Ceph Jika Storage Backend Down

```bash
# Cek NFS server bisa di-reach
showmount -e <nfs-server-ip>

# Cek export directory
ls /exports/data

# Test mount manual dari node
ssh k3s-node "sudo mount -t nfs <nfs-server>:/exports/data /mnt/test"
```

---

## 🛡️ Prevention

### Prevention 1 — Helm/Kustomize untuk Bundle PVC + App

```yaml
# charts/app/templates/pvc.yaml
{{- if .Values.persistence.enabled }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "app.fullname" . }}
  namespace: {{ .Release.Namespace }}
spec:
  accessModes:
  - {{ .Values.persistence.accessMode | default "ReadWriteOnce" }}
  resources:
    requests:
      storage: {{ .Values.persistence.size | default "1Gi" }}
  storageClassName: {{ .Values.persistence.storageClass | default "local-path" }}
{{- end }}
```

**Keuntungan:** PVC dan Deployment di-bundle dalam 1 chart. Saat install chart, PVC otomatis dibuat.

### Prevention 2 — GitOps Policy untuk Validasi Reference

```yaml
# Kyverno policy: enforce PVC exists before Deployment
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-pvc
spec:
  validationFailureAction: Enforce
  rules:
  - name: check-pvc-reference
    match:
      any:
      - resources:
          kinds:
          - Deployment
          - StatefulSet
    validate:
      message: "All volumeClaimTemplates must have matching PVC"
      pattern:
        spec:
          template:
            spec:
              volumes:
              - persistentVolumeClaim:
                  claimName: "?*"
```

### Prevention 3 — PVC Naming Convention

```
{namespace}-{app-name}-{purpose}-{env}
Contoh: produksi-go-app-data-prod
```

Dokumentasikan di `docs/storage-convention.md` di repo.

### Prevention 4 — Backup dengan Velero

```bash
# Install Velero (https://velero.io)
velero install   --provider aws   --bucket produksi-backup   --prefix k8s-backup   --secret-file ./credentials-velero

# Backup harian otomatis
velero schedule create daily-backup   --schedule="0 2 * * *"   --include-namespaces produksi,staging

# Restore saat PVC rusak
velero restore create --from-backup daily-backup-20260810020000
```

### Prevention 5 — Pre-flight Check di CI

```bash
# .gitlab-ci.yml
validate-references:
  stage: validate
  script:
  - |
    # Extract semua PVC yang direference
    REFERENCED=$(yq eval-all '.. | select(has("persistentVolumeClaim")).persistentVolumeClaim.claimName' k8s/*.yaml | sort -u)
    
    # Extract semua PVC yang didefinisikan
    DEFINED=$(yq eval-all 'select(.kind == "PersistentVolumeClaim") | .metadata.name' k8s/*.yaml | sort -u)
    
    # Cek referenced ⊆ defined
    for pvc in $REFERENCED; do
      if ! echo "$DEFINED" | grep -q "^$pvc$"; then
        echo "❌ PVC '$pvc' referenced but not defined"
        exit 1
      fi
    done
    
    echo "✅ All PVC references valid"
```

### Prevention 6 — ResourceQuota untuk PVC per Namespace

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: insiden-lab-storage
  namespace: insiden-lab
spec:
  hard:
    persistentvolumeclaims: "10"
    requests.storage: "20Gi"
```

**Efek:** Membatasi jumlah dan total storage di namespace — mencegah overuse.

---

## 🧪 Verifikasi Recovery

```bash
# 1. Pod Running
kubectl get pods -n insiden-lab -l app=failedmount-app
# NAME                              READY   STATUS    RESTARTS   AGE
# failedmount-app-7f9c8d9f8-abcd1   1/1     Running   0          30s

# 2. PVC Bound
kubectl get pvc -n insiden-lab
# NAME                  STATUS   VOLUME                                     CAPACITY
# data-pvc-tidak-ada    Bound    pvc-abc123-xxx                             1Gi

# 3. Mount point bisa ditulis
kubectl exec -n insiden-lab <pod-name> -- ls -la /data
# Output: drwxr-xr-x 2 root root 4096 Aug 10 09:50 .
#         drwxr-xr-x 2 root root 4096 Aug 10 09:50 ..

kubectl exec -n insiden-lab <pod-name> -- touch /data/test.txt
kubectl exec -n insiden-lab <pod-name> -- ls /data
# Output: test.txt    ← berhasil dibuat

# 4. Alert resolved
# Cek Slack: [RESOLVED] PodsFailedMount
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/05-failedmount-missing-pvc.yaml
kubectl delete pvc data-pvc-tidak-ada -n insiden-lab
```

---

## 📖 Rangkuman

| Aspek | Catatan |
|---|---|
| **Symptom utama** | Stuck `ContainerCreating`, Events: `FailedMount` |
| **Cara diagnosis** | `kubectl describe pod` → Events → lihat message FailedMount |
| **Penyebab umum** | PVC missing, StorageClass invalid, permission issue |
| **Mitigation** | Buat PVC, fix SC name, hapus volume ref |
| **Prevention terbaik** | Helm bundle PVC + GitOps validation + Velero backup |
| **Alert yang firing** | `PodsFailedMount` (Minggu 8) ✅ |

---

## 🎓 Penutup Minggu 9

Selamat! Kamu sudah menguasai **5 incident paling umum** di Kubernetes:

| Incident | Exit/Status | Root Cause Umum |
|---|---|---|
| **CrashLoopBackOff** | Exit 1 | Bug code, missing env var |
| **OOMKilled** | Exit 137 | Memory limit terlalu rendah / leak |
| **Pending Pod** | Pending | Resource / nodeSelector issue |
| **ImagePullBackOff** | ImagePullBackOff | Tag typo / no credentials |
| **FailedMount** | ContainerCreating | PVC missing / wrong SC |

**Skill yang sudah kamu dapat:**

1. ✅ Membaca `kubectl describe pod` dengan percaya diri
2. ✅ Mengenali exit code (1, 137, 143) dan artinya
3. ✅ Menggunakan 5 Whys untuk root cause analysis
4. ✅ Memilih mitigation yang tepat
5. ✅ Mendesain prevention (probe, quota, validation, backup)

**Hubungan dengan observability stack:**

```
Metrics (M5) → deteksi resource usage anomaly
Logs (M6)    → cari error message spesifik
Traces (M7)  → lihat bottleneck latency
Alerts (M8)  → otomatis firing saat incident
Incident (M9) → GUNAKAN semua itu untuk troubleshoot!
```

**Minggu depan** (Minggu 10) kita akan naik level ke **insiden yang lebih kompleks** — yang root cause-nya tidak langsung terlihat, butuh korelasi data dari Metrics + Logs + Traces + kubectl:

- CPU Spike (mungkin due to GC? atau attack? atau legitimate traffic?)
- Memory Leak (butuh heap dump analysis)
- Disk Full
- DNS Error
- PVC Full
- Network Timeout
- Latency
- Slow Database
- Deadlock

👉 Lanjut ke `minggu-10/README.md`
