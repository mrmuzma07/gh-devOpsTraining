# Minggu 9 — Modul 04: Incident Pending Pod

> **"Pod baru di-apply tapi tidak jalan-jalan, statusnya Pending selamanya."**

Pod Pending artinya K8s **tidak bisa menjadwalkan pod** ke node mana pun. Penyebabnya bisa:

1. **Resource tidak cukup** — pod minta lebih banyak CPU/memory dari yang tersedia di cluster
2. **Selector tidak match** — nodeSelector, affinity, atau tolerations tidak puas
3. **PVC tidak tersedia** — pod butuh volume tapi PVC belum di-bind
4. **Node dalam kondisi NotReady** — semua node yang cocok sedang down

Analogi: seperti kamu pesan taksi online, tapi tidak ada driver yang mau ambil order karena:
- Alamat terlalu jauh (resource tidak cukup)
- Tipe mobil yang kamu minta tidak ada (nodeSelector)
- Semua driver sedang off (NotReady node)

---

## 🎯 Tujuan Modul

1. Membedakan **5 jenis Pending** berdasarkan event message
2. Membaca **node capacity & allocatable** dengan kubectl
3. Memahami **nodeSelector, affinity, taints/tolerations**
4. Memilih **mitigation** yang tepat (kurangi request, tambah node, hapus selector)
5. Menerapkan **ResourceQuota, LimitRange, PodDisruptionBudget** untuk prevention

---

## 📦 Simulasi 1 — Pending Karena Resource Overcommit

Pod minta CPU/memory yang tidak masuk akal.

### File: `manifests/03-pending-overcommit.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pending-app
  namespace: insiden-lab
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pending-app
  template:
    metadata:
      labels:
        app: pending-app
    spec:
      containers:
      - name: app
        image: nginx:1.27-alpine
        resources:
          requests:
            cpu: "16"
            memory: 64Gi
          limits:
            cpu: "16"
            memory: 64Gi
        ports:
        - containerPort: 80
```

### Deploy & Observe

```bash
kubectl apply -f manifests/03-pending-overcommit.yaml

kubectl get pods -n insiden-lab -l app=pending-app -w
```

**Output:**
```
NAME                          READY   STATUS    RESTARTS   AGE
pending-app-7f9c8d9f8-abcd1   0/1     Pending   0          5s
pending-app-7f9c8d9f8-efgh2   0/1     Pending   0          5s
```

**Tidak akan pernah jadi Running** — di laptop dengan 8 CPU, minta 16 CPU jelas mustahil.

---

## 📦 Simulasi 2 — Pending Karena nodeSelector Tidak Match

### File: `manifests/03b-pending-node-selector.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-app
  namespace: insiden-lab
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gpu-app
  template:
    metadata:
      labels:
        app: gpu-app
    spec:
      nodeSelector:
        hardware: nvidia-a100    # tidak ada node dengan label ini
      containers:
      - name: app
        image: nginx:1.27-alpine
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
```

### Deploy & Observe

```bash
kubectl apply -f manifests/03b-pending-node-selector.yaml

kubectl get pods -n insiden-lab -l app=gpu-app
# NAME                      READY   STATUS    RESTARTS   AGE
# gpu-app-xxx               0/1     Pending   0          30s
```

---

## 🔍 Symptoms

### Symptom 1 — Pod Stuck di `Pending`

```bash
$ kubectl get pods -n insiden-lab
NAME                          READY   STATUS    RESTARTS   AGE
pending-app-7f9c8d9f8-abcd1   0/1     Pending   0          5m     ← stuck 5 menit
```

**Tanda chronic:** Pod Pending > 1 menit = ada masalah scheduling.

### Symptom 2 — Alert firing (kamu setup di Minggu 8)

```
[FIRING] PodsPendingTooLong
  namespace = insiden-lab
  pod       = pending-app-xxx
  pending   = 5 minutes
  severity  = warning
```

### Symptom 3 — Endpoints kosong

```bash
$ kubectl get endpoints pending-app -n insiden-lab
NAME           ENDPOINTS
pending-app    <none>     ← tidak ada IP karena pod belum jalan
```

---

## 🕵️ Investigation

### Step 1 — describe pod (lihat Events)

```bash
$ kubectl describe pod pending-app-7f9c8d9f8-abcd1 -n insiden-lab
```

**Output penting:**

```
Events:
  Type     Reason             Age   From              Message
  ----     ------             ----  ----              -------
  Warning  FailedScheduling   3m    default-scheduler  0/3 nodes are available:
                                     insufficient cpu (3).
                                     insufficient memory (3).
```

**PEMBACAAN:**
- `0/3 nodes are available` → scheduler cek 3 node, tidak ada yang lolos
- `insufficient cpu (3)` → semua 3 node CPU-nya tidak cukup
- `insufficient memory (3)` → semua 3 node memory-nya tidak cukup

### Step 2 — Lihat capacity node

```bash
$ kubectl get nodes -o custom-columns=  NAME:.metadata.name,  CPU:.status.allocatable.cpu,  MEM:.status.allocatable.memory,  PODS:.status.allocatable.pods
```

**Output (contoh):**
```
NAME          CPU    MEM        PODS
laptop-k3s    8      16Gi       30
```

**Artinya:**
- Total CPU cluster: 8 core
- Total memory cluster: 16 GB
- Maks 30 pod total

Pod kita minta 16 CPU × 2 replica = 32 CPU → **mustahil** di cluster ini.

### Step 3 — Lihat resource yang sudah dialokasi

```bash
$ kubectl describe node laptop-k3s | grep -A 10 "Allocated"
```

```
Allocated resources:
  CPU Requests:    450m (5%)      ← hanya 5% terpakai
  CPU Limits:      1 (12%)
  Memory Requests: 800Mi (4%)
  Memory Limits:   2Gi (12%)
```

**Kesimpulan:** Cluster masih banyak sisa resource. Jadi **bukan karena resource cluster penuh**, tapi karena pod minta **terlalu banyak** untuk 1 replica.

### Step 4 — Untuk kasus nodeSelector

```bash
$ kubectl describe pod gpu-app-xxx -n insiden-lab | tail -10
```

**Output:**
```
Events:
  Type     Reason             Age   From               Message
  ----     ------             ----  ----               -------
  Warning  FailedScheduling   30s   default-scheduler  0/1 nodes are available:
                                      1 node(s) didn't match Pod's node selector.
```

**PEMBACAAN:** Scheduler menemukan 1 node, tapi `nodeSelector` tidak match. Solusi: tambah label di node, atau hapus nodeSelector.

### Step 5 — Lihat label semua node

```bash
$ kubectl get nodes --show-labels
```

**Output:**
```
NAME          STATUS   ROLES                  AGE     VERSION   LABELS
laptop-k3s    Ready    control-plane,master   30d     v1.30.4   beta.kubernetes.io/arch=amd64,
                                                            kubernetes.io/hostname=laptop-k3s,
                                                            node-role.kubernetes.io/control-plane=true,
                                                            node-role.kubernetes.io/master=true
```

Tidak ada label `hardware=nvidia-a100` → sesuai dugaan.

---

## 🎯 Root Cause Analysis

### 5 Whys untuk Resource Overcommit

```
Problem: Pod pending-app Pending 5 menit

  Why 1: Kenapa Pending?
    → Scheduler tidak bisa assign ke node manapun
  
  Why 2: Kenapa scheduler tidak bisa assign?
    → semua node insufficient cpu & memory
  
  Why 3: Kenapa insufficient?
    → pod minta 16 CPU + 64Gi, node cuma punya 8 CPU + 16Gi
  
  Why 4: Kenapa request-nya tidak realistis?
    → Developer copy-paste dari spec production
  
  Why 5: Kenapa production spec tidak valid?
    → Tidak pernah di-test di staging environment

🎯 ROOT CAUSE: Resource request tidak realistis untuk cluster ini
```

### Tabel Diagnosis Pending Berdasarkan Event

| Event Message | Root Cause | Solusi |
|---|---|---|
| `0/N nodes are available: insufficient cpu` | CPU request > capacity | Kurangi request, tambah node |
| `0/N nodes are available: insufficient memory` | Memory request > capacity | Kurangi request, tambah node |
| `N node(s) didn't match Pod's node selector` | Selector tidak match | Hapus/ubah selector |
| `N node(s) had taint that the pod didn't tolerate` | Taints tidak ditolerate | Tambah toleration |
| `N node(s) had volume node affinity conflict` | Volume affinity mismatch | Ubah topology constraint |
| `N node(s) were not ready` | Semua node down | Recovery node |
| `persistentvolumeclaim "xxx" not found` | PVC missing | Buat PVC atau hapus reference |
| `waiting for first consumer to be created before binding` | PVC mode WaitForFirstConsumer | Normal untuk dynamic provision |

---

## 🛠️ Mitigation

### Opsi A — Kurangi Resource Request

```bash
# Patch deployment dengan request realistis
kubectl set resources deployment/pending-app -n insiden-lab   --requests=cpu=100m,memory=128Mi   --limits=cpu=500m,memory=256Mi

# Tunggu scheduler retry (biasanya dalam 1-2 menit)
kubectl get pods -n insiden-lab -l app=pending-app -w
```

### Opsi B — Kurangi Replicas

```bash
# Kalau 2 replica tidak muat, coba 1
kubectl scale deployment/pending-app -n insiden-lab --replicas=1
```

### Opsi C — Tambah Node (untuk multi-node cluster)

```bash
# Untuk k3s: jalankan k3s agent di node kedua
curl -sfL https://get.k3s.io | K3S_URL=https://<server-ip>:6443   K3S_TOKEN=<shared-secret> sh -
```

### Opsi D — Hapus/ubah nodeSelector

```bash
# Hapus nodeSelector via patch
kubectl patch deployment/gpu-app -n insiden-lab --type=json   -p='[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'

# Atau tambahkan label di node
kubectl label node laptop-k3s hardware=nvidia-a100
# (HANYA untuk simulasi, jangan labeli node dengan GPU yang tidak ada!)
```

### Opsi E — Buat PVC yang Hilang (untuk PVC-related Pending)

```bash
# Lihat PVC yang dibutuhkan
kubectl describe pod <pod> -n <ns> | grep -A 2 "Events"
# Output: persistentvolumeclaim "data-pvc" not found

# Buat PVC-nya
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: insiden-lab
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Gi
EOF
```

---

## 🛡️ Prevention

### Prevention 1 — ResourceQuota per Namespace

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: insiden-lab-quota
  namespace: insiden-lab
spec:
  hard:
    pods: "20"           # max 20 pod di namespace ini
    requests.cpu: "4"    # total CPU request max 4 core
    requests.memory: 8Gi # total memory request max 8 GB
    limits.cpu: "8"
    limits.memory: 16Gi
```

**Efek:** Kalau developer coba apply pod dengan request berlebihan, **akan ditolak oleh API server** sebelum sempat Pending.

### Prevention 2 — LimitRange (default per pod)

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: insiden-lab-limits
  namespace: insiden-lab
spec:
  limits:
  - type: Container
    default:           # default limit kalau developer tidak set
      cpu: 500m
      memory: 512Mi
    defaultRequest:    # default request
      cpu: 100m
      memory: 128Mi
    max:               # max yang boleh di-request
      cpu: "2"
      memory: 4Gi
    min:               # min yang boleh di-request
      cpu: 50m
      memory: 64Mi
```

### Prevention 3 — Node Affinity Documentation

Pastikan semua nodeSelector / affinity di dokumentasikan di repo:

```markdown
# Hardware Labels

| Label | Node Type | Resource |
|-------|-----------|----------|
| workload=general | laptop-k3s | 8 CPU / 16 GB |
| workload=gpu | (none in dev) | 16 CPU / 64 GB + GPU |
| workload=memory | (none in dev) | 16 CPU / 256 GB RAM |

## Pod Placement Rules

| Service | nodeSelector | Reason |
|---------|--------------|--------|
| go-app | none | General workload |
| ml-train | workload=gpu | Butuh GPU |
| redis-cache | workload=memory | Butuh RAM besar |
```

### Prevention 4 — PodDisruptionBudget (untuk availability)

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: go-app-pdb
  namespace: insiden-lab
spec:
  minAvailable: 1     # min 1 pod harus running saat disruption
  selector:
    matchLabels:
      app: go-app
```

**Catatan:** PDB tidak mencegah Pending, tapi mencegah **pod hilang** saat node maintenance. Penting untuk availability.

### Prevention 5 — Pre-flight Check di CI/CD

```bash
# Di .gitlab-ci.yml
deploy:
  script:
  - |
    # Cek apakah manifest request valid
    total_cpu=$(yq eval-all '[.. | .resources?.requests?.cpu? // "0"] | map(.| sub("m","") | tonumber/1000) | add' k8s/*.yaml)
    node_cpu=$(kubectl get nodes -o jsonpath='{.items[*].status.allocatable.cpu}' | tr ' ' '
' | sed 's/m$//' | awk '{s+=$1} END {print s/1000}')
    
    if (( $(echo "$total_cpu > $node_cpu" | bc -l) )); then
      echo "ERROR: Total CPU request ($total_cpu) > cluster capacity ($node_cpu)"
      exit 1
    fi
```

---

## 🧪 Verifikasi Recovery

```bash
# 1. Pod Running
kubectl get pods -n insiden-lab -l app=pending-app
# NAME                          READY   STATUS    RESTARTS   AGE
# pending-app-xxx               1/1     Running   0          30s

# 2. Scheduled ke node
kubectl get pods -n insiden-lab -l app=pending-app   -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName
# NAME                NODE
# pending-app-xxx     laptop-k3s

# 3. Alert resolved
# Cek Slack: [RESOLVED] PodsPendingTooLong
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/03-pending-overcommit.yaml
kubectl delete -f manifests/03b-pending-node-selector.yaml
```

---

## 📖 Rangkuman

| Aspek | Catatan |
|---|---|
| **Symptom utama** | STATUS=Pending, tidak pernah Running |
| **Cara diagnosis** | `kubectl describe pod` → Events → FailedScheduling message |
| **5 jenis Pending umum** | resource, selector, taints, PVC, NotReady node |
| **Mitigation** | Kurangi request, tambah node, hapus selector |
| **Prevention terbaik** | ResourceQuota + LimitRange + dokumentasi placement |
| **Alert yang firing** | `PodsPendingTooLong` (Minggu 8) ✅ |

---

## ➡️ Modul 05: ImagePullBackOff

Pending adalah masalah **scheduling** (resource/placement). Sekarang kita masuk masalah **image pulling** — pod tidak bisa start karena K8s tidak bisa download container image-nya.

👉 Lanjut ke `05-incident-imagepullbackoff.md`
