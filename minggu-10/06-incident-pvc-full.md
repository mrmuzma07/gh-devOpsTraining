# Incident #6 — Persistent Volume Claim (PVC) Capacity Exhaustion

> **Satu kalimat:** Ketika disk Persistent Volume (PV) yang digunakan oleh Database atau Pod Stateful penuh 100%, **aplikasi akan mengalami CrashLoopBackOff, database Read-Only, atau transaksi gagal total** akibat ketidakmampuan menulis data baru (`no space left on device`).

Berbeda dengan Ephemeral Storage (yang menempel pada Lifecycle Pod/Node), **Persistent Volume (PV)** digunakan untuk menyimpan data jangka panjang seperti PostgreSQL, MySQL, Redis AOF, atau Log Storage. Jika PVC penuh, aplikasi tidak bisa sekadar di-restart karena data lama berada di volume tersebut.

---

## 🎯 Learning Outcomes

Setelah menyelesaikan insiden ini, Anda mampu:
1. Memahami perbedaan **Volume Expansion** (Resizing PVC) vs **Re-provisioning StorageClass**.
2. Mengidentifikasi PVC yang kapasitasnya menipis sebelum terjadinya crash total.
3. Melakukan **Online PVC Expansion** tanpa Downtime aplikasi di Kubernetes (CSI Storage Driver).
4. Melakukan pembersihan data / vacuum darurat saat storage 100% penuh.
5. Memonitor metrics storage PVC lewat Prometheus / Mimir (`kubelet_volume_stats_used_bytes`).

---

## 1. 🩺 Gejala (Symptoms)

### 1.1 Dari Sisi Aplikasi (Logs & State)
- PostgreSQL Pod berubah state menjadi Read-Only atau Crash:
  ```text
  PANIC: could not write to file "pg_wal/xlogtemp.19283": No space left on device
  FATAL: terminating connection due to administrator command
  ```
- Application Web API gagal menyimpan data (HTTP 500):
  ```text
  Internal Error: SQLSTATE[HY000]: General error: 1021 Disk full (/tmp/#sql-128a_1.MAI); waiting for someone to free some space...
  ```

### 1.2 Dari Sisi Kubernetes (`kubectl`)
```bash
$ kubectl get pvc -n database
NAME             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
postgres-data    Bound    pvc-87123912-3819-411a-a123-9821a8123912   10Gi       RWO            local-path     45d

$ kubectl get pod -n database
NAME         READY   STATUS             RESTARTS   AGE
postgres-0   0/1     CrashLoopBackOff   8          30m
```

Event Pod menunjukkan error mounting atau I/O write failures:
```bash
$ kubectl describe pod postgres-0 -n database
Events:
  Warning  BackOff  2m (x12 over 8m)  kubelet  Back-off restarting failed container
```

### 1.3 Dari Sisi Mimir & Alertmanager
- Alert Critical: `KubePersistentVolumeFillingUp` atau `KubePersistentVolumeFull`.
- PromQL Metric:
```promql
# Persentase penggunaan storage PVC
(kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) * 100 > 85
# Hasil: postgres-data-postgres-0 = 99.8% !
```

---

## 2. 🔍 Investigasi Step-by-Step

### Step 1 — Periksa Penggunaan Disk dari Metrics Kubelet
Metrik PVC diukur secara akurat oleh `kubelet` melalui plugin storage:
```bash
$ kubectl get --raw /api/v1/nodes/laptop-node-1/proxy/stats/summary | jq '.pods[].volume[]? | select(.name=="postgres-data")'
{
  "name": "postgres-data",
  "capacityBytes": 107374182400,
  "usedBytes": 107370000000,
  "availableBytes": 4182400,
  "percentageUsed": 99.96
}
```

### Step 2 — Verifikasi Kemampuan StorageClass (Can Expand?)
Sebelum mencoba memperbesar PVC, pastikan `StorageClass` yang digunakan mendukung fiturnya:
```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION
local-path (default) rancher.io/local-path   Delete          WaitForFirstConsumer   true
```

> **Kunci:** Kolom `ALLOWVOLUMEEXPANSION` **HARUS `true`**. Jika `false`, Anda tidak bisa langsung mengubah ukuran PVC.

---

## 3. 🛠️ Penanganan & Mitigasi (Step-by-Step Recovery)

### Skenario A: StorageClass Mendukung Volume Expansion (`ALLOWVOLUMEEXPANSION: true`)

Ini adalah metode **Zero Downtime / Minimal Intervention**.

#### Step 1 — Edit PVC Capacity
Ubah kapasitas PVC dari 10Gi menjadi 20Gi:
```bash
$ kubectl edit pvc postgres-data -n database
```
Atau patch menggunakan kubectl CLI:
```bash
$ kubectl patch pvc postgres-data -n database -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'
persistentvolumeclaim/postgres-data patched
```

#### Step 2 — Verifikasi Status Resizing
```bash
$ kubectl get pvc postgres-data -n database
NAME            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
postgres-data   Bound    pvc-87123912-3819-411a-a123-9821a8123912   20Gi       RWO            local-path     45d
```

> **Catatan Tambahan (File System Expansion):**
> Pada beberapa CSI Driver (seperti AWS EBS, Longhorn, Ceph), PVC capacity di etcd langsung berubah 20Gi, tetapi filesystem (ext4/xfs) di dalam PV butuh penyesuaian oleh Kubelet. Kubelet akan otomatis melakukan `resize2fs` atau `xfs_growfs` ketika Pod dalam keadaan `Running` atau di-restart sekali.

---

### Skenario B: StorageClass TIDAK Mendukung Volume Expansion (`ALLOWVOLUMEEXPANSION: false`)

Jika StorageClass tidak mengizinkan resize langsung, gunakan langkah emergency berikut:

#### Step 1 — Jalankan Temporary Pod untuk Bersihkan Disk (Emergency Maintenance)
Jika database crash dan tidak mau up karena disk 100% penuh:
1. Scale down StatefulSet/Deployment ke 0:
   ```bash
   $ kubectl scale statefulset postgres -n database --replicas=0
   ```
2. Buat Maintenance Pod sederhana yang menempel ke PVC tersebut:
   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: storage-fixer
     namespace: database
   spec:
     containers:
     - name: fixer
       image: alpine:3.19
       command: ["tail", "-f", "/dev/null"]
       volumeMounts:
       - name: data
         mountPath: /mnt/data
     volumes:
     - name: data
       persistentVolumeClaim:
         claimName: postgres-data
   ```
3. Exec ke Pod maintenance untuk menghapus WAL log tua atau file dump temporer:
   ```bash
   $ kubectl exec -it storage-fixer -n database -- /bin/sh
   / # cd /mnt/data
   /mnt/data # du -sh * | sort -hr
   /mnt/data # rm -rf pg_wal/old_wal_file.log  # Hapus file aman / log lama!
   ```
4. Hapus Pod maintenance dan Scale-Up kembali StatefulSet original:
   ```bash
   $ kubectl delete pod storage-fixer -n database
   $ kubectl scale statefulset postgres -n database --replicas=1
   ```

---

## 4. 🛡️ Strategi Pencegahan Jangka Panjang

### 4.1 Prometheus Alerting Rules untuk PVC (Alert Dini)
Pasang alert yang tidak hanya melihat persentase, tetapi juga **prediksi kecepatan disk terisi (Linear Prediction)**:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: pvc-capacity-alerts
  namespace: monitoring
spec:
  groups:
  - name: pvc-alerts
    rules:
    # Alert 1: Kapasitas saat ini > 85%
    - alert: PersistentVolumeFillingUp
      expr: |
        (kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) * 100 > 85
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "PVC {{ $labels.persistentvolumeclaim }} di namespace {{ $labels.namespace }} terpakai {{ $value | printf \"%.1f\" }}%"

    # Alert 2: Prediksi disk akan penuh dalam 24 Jam berdasarkan trend 4 jam terakhir
    - alert: PersistentVolumeWillFillIn24Hours
      expr: |
        predict_linear(kubelet_volume_stats_available_bytes[4h], 24 * 3600) < 0
      for: 15m
      labels:
        severity: critical
      annotations:
        summary: "PVC {{ $labels.persistentvolumeclaim }} diperkirakan HABIS dalam 24 Jam!"
```

### 4.2 Auto-Scaling PVC (KEDA / Resizer Operator)
Gunakan Controller tambahan seperti `kube-pv-exporter` atau `kube-pvc-autoresizer` yang secara otomatis menaikkan kuota PVC ketika mencapai threshold 80%.

---

## 5. 🧪 Lab Hands-On: Simulasi PVC Penuh & Online Expansion

### Step 1 — Deploy PVC & Stateful Pod
File: `minggu-10/manifests/06-pvc-lab.yaml`
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: demo-pvc-expansion
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: pvc-writer
  namespace: default
spec:
  containers:
  - name: writer
    image: alpine:3.19
    command: ["/bin/sh", "-c"]
    args:
    - |
      echo "Mengisi PVC..."
      dd if=/dev/zero of=/data/fill.bin bs=1M count=900
      echo "Selesai mengisi 900MB data."
      sleep 3600
    volumeMounts:
    - name: store
      mountPath: /data
  volumes:
  - name: store
    persistentVolumeClaim:
      claimName: demo-pvc-expansion
```

### Step 2 — Jalankan & Amati Kondisi Penuh
```bash
$ kubectl apply -f manifests/06-pvc-lab.yaml
$ kubectl exec -it pvc-writer -- df -h /data
Filesystem      Size  Used Avail Use% Mounted on
/dev/sdb        970M  900M   70M  93% /data
```

### Step 3 — Eksekusi Resizing Live
```bash
$ kubectl patch pvc demo-pvc-expansion -p '{"spec":{"resources":{"requests":{"storage":"3Gi"}}}}'
persistentvolumeclaim/demo-pvc-expansion patched

# Cek hasil perubahan langsung di Pod tanpa restart!
$ kubectl exec -it pvc-writer -- df -h /data
Filesystem      Size  Used Avail Use% Mounted on
/dev/sdb        2.9G  900M  2.0G  30% /data   # KAPASITAS BERHASIL NAIK 3GB!
```

---

## 6. 📋 Cheat Sheet Quick-Reference

```text
AKSI / PROBLEM                     | COMMAND / SOLUSI
-----------------------------------|-----------------------------------------------------------------
Cek PVC terpakai di cluster        | kubectl get pvc -A
Cek ketersediaan Volume Expansion  | kubectl get sc -o jsonpath='{.items[*].allowVolumeExpansion}'
Patch / Enlarge PVC Size           | kubectl patch pvc <name> -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'
Emergency Clean-up                | Scale StatefulSet to 0 -> Spin Alpine pod -> rm unwanted files
Prediksi Disk Habis (PromQL)       | predict_linear(kubelet_volume_stats_available_bytes[4h], 86400) < 0
```

---

**Lanjut ke insiden berikutnya:** [07-incident-network-timeout.md](./07-incident-network-timeout.md) — saat komunikasi antar Pod mengalami Network Connection Timeout akibat NetworkPolicy atau CNI plugin issue.
