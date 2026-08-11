# Minggu 9 — Modul 05: Incident ImagePullBackOff

> **"Pod baru di-deploy, tapi statusnya ImagePullBackOff — tidak pernah jalan."**

ImagePullBackOff artinya K8s (lebih spesifik: **kubelet** di node) **gagal download container image** dari registry. Ini terjadi SEBELUM container dijalankan — pod bahkan tidak masuk fase `Pending: ContainerCreating`, melainkan langsung `Waiting: ImagePullBackOff`.

Penyebab umum:

1. **Image tag salah** — typo atau tag tidak ada di registry
2. **Private registry tanpa credentials** — image butuh login tapi tidak ada Secret
3. **Registry tidak reachable** — DNS/network issue
4. **Rate limit registry** — Docker Hub membatasi request anonymous

Analogi: seperti kamu pesan barang di toko online, tapi alamatnya salah ketik (Image not found), atau toko-nya butuh会员カード (no credentials), atau toko-nya tutup (registry down).

---

## 🎯 Tujuan Modul

1. Membedakan 4 jenis error pulling image dari event message
2. Membaca **imagePullSecrets** dan konfigurasinya
3. Membuat **docker-registry Secret** untuk private registry
4. Memilih **mitigation** yang tepat (fix tag, create secret, fix network)
5. Menerapkan **image scanning + verification** di CI/CD sebagai prevention

---

## 📦 Simulasi 1 — ImagePullBackOff Karena Tag Salah

### File: `manifests/04-imagepull-bad-tag.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: imagepull-app
  namespace: insiden-lab
spec:
  replicas: 1
  selector:
    matchLabels:
      app: imagepull-app
  template:
    metadata:
      labels:
        app: imagepull-app
    spec:
      containers:
      - name: app
        image: nginx:v999.999.999-typo    # ← TAG INI TIDAK ADA
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
```

### Deploy & Observe

```bash
kubectl apply -f manifests/04-imagepull-bad-tag.yaml

kubectl get pods -n insiden-lab -l app=imagepull-app -w
```

**Output:**
```
NAME                            READY   STATUS                 RESTARTS   AGE
imagepull-app-7f9c8d9f8-abcd1   0/1     ErrImagePull           0          5s
imagepull-app-7f9c8d9f8-abcd1   0/1     ImagePullBackOff       0          30s
```

**Catatan:** Pertama muncul `ErrImagePull` (immediate attempt), lalu `ImagePullBackOff` (setelah retry dengan backoff).

---

## 📦 Simulasi 2 — ImagePullBackOff Karena Private Registry Tanpa Secret

### File: `manifests/04b-imagepull-private-no-secret.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: private-app
  namespace: insiden-lab
spec:
  replicas: 1
  selector:
    matchLabels:
      app: private-app
  template:
    metadata:
      labels:
        app: private-app
    spec:
      containers:
      - name: app
        image: ghcr.io/perusahaan-rahasia/secret-app:v1.0.0
        # TIDAK ADA imagePullSecrets → anonymous access ditolak
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
```

### Deploy & Observe

```bash
kubectl apply -f manifests/04b-imagepull-private-no-secret.yaml

kubectl get pods -n insiden-lab -l app=private-app
# NAME                          READY   STATUS             RESTARTS   AGE
# private-app-xxx               0/1     ImagePullBackOff   0          30s
```

---

## 🔍 Symptoms

### Symptom 1 — Status `ErrImagePull` atau `ImagePullBackOff`

```bash
$ kubectl get pods -n insiden-lab
NAME                            READY   STATUS             RESTARTS   AGE
imagepull-app-7f9c8d9f8-abcd1   0/1     ImagePullBackOff   0          3m
```

### Symptom 2 — Alert firing

```
[FIRING] ImagePullBackOff
  namespace = insiden-lab
  pod       = imagepull-app-xxx
  image     = nginx:v999.999.999-typo
  reason    = pull access denied / not found
  severity  = warning
```

### Symptom 3 — Image tidak ter-cache di node

```bash
$ kubectl get pods -n insiden-lab -l app=imagepull-app     -o jsonpath='{.items[0].status.containerStatuses[0].state.waiting.message}'
# Output: Back-off pulling image "nginx:v999.999.999-typo"
```

---

## 🕵️ Investigation

### Step 1 — describe pod

```bash
$ kubectl describe pod imagepull-app-7f9c8d9f8-abcd1 -n insiden-lab
```

**Output penting:**

```
Containers:
  app:
    Container ID:  
    Image:         nginx:v999.999.999-typo
    State:         Waiting
      Reason:      ImagePullBackOff
    Ready:         False
    Restart Count: 0

Events:
  Type     Reason          Age              From     Message
  ----     ------          ----             ----     -------
  Normal   Scheduled       30s              default-scheduler  Successfully assigned ...
  Normal   Pulling         20s              kubelet   Pulling image "nginx:v999.999.999-typo"
  Warning  Failed          15s              kubelet   Failed to pull image "nginx:v999.999.999-typo":
                                                   rpc error: code = NotFound
                                                   desc = failed to pull and unpack image
                                                   "docker.io/library/nginx:v999.999.999-typo":
                                                   failed to resolve reference
                                                   "docker.io/library/nginx:v999.999.999-typo":
                                                   failed to authorize: failed to fetch oauth token:
                                                   404 Not Found
  Warning  Failed          10s              kubelet   Error: ErrImagePull
  Normal   BackOff          5s              kubelet   Back-off pulling image "nginx:v999.999.999-typo"
```

**Pesan kunci:**
- `failed to resolve reference` → image reference tidak valid
- `failed to fetch oauth token: 404 Not Found` → registry merespons 404 (tag tidak ada)

### Step 2 — Verifikasi tag di registry

```bash
# Cek apakah tag ada di Docker Hub
curl -s https://hub.docker.com/v2/repositories/library/nginx/tags/v999.999.999-typo/ | head
# Output: {"error":"not found"}

# Cek tag yang valid
curl -s https://hub.docker.com/v2/repositories/library/nginx/tags/1.27-alpine/ | head
# Output: {"name":"1.27-alpine",...}
```

### Step 3 — Untuk kasus private registry

```bash
$ kubectl describe pod private-app-xxx -n insiden-lab | tail -20
```

**Output:**
```
Events:
  Warning  Failed       10s   kubelet   Failed to pull image "ghcr.io/perusahaan-rahasia/secret-app:v1.0.0":
                                       rpc error: code = Unknown
                                       desc = failed to pull and unpack image
                                       "ghcr.io/perusahaan-rahasia/secret-app:v1.0.0":
                                       failed to resolve reference
                                       "ghcr.io/perusahaan-rahasia/secret-app:v1.0.0":
                                       pull access denied, repository does not exist
                                       or may require 'docker login': denied
```

**Pesan kunci:**
- `pull access denied` → registry butuh auth
- `may require docker login` → tidak ada Secret

### Step 4 — Cross-check dengan crictl di node

```bash
# Login ke node k3s (di k3s, node = laptop sendiri)
sudo crictl pull nginx:v999.999.999-typo
# Output: FATA[0000] pulling image failed: rpc error:
#         code = NotFound desc = failed to pull ...

# Lihat images yang sudah ter-cache
sudo crictl images
```

### Step 5 — Cek konfigurasi image registry

```bash
# Cek apakah ada global registry mirror (untuk k3s)
cat /etc/rancher/k3s/registries.yaml 2>/dev/null
# Contoh output:
# mirrors:
#   docker.io:
#     endpoint:
#       - https://registry-mirror.internal:5000
```

---

## 🎯 Root Cause Analysis

### 5 Whys untuk Tag Salah

```
Problem: Pod imagepull-app ImagePullBackOff

  Why 1: Kenapa ImagePullBackOff?
    → K8s gagal pull image "nginx:v999.999.999-typo"
  
  Why 2: Kenapa gagal pull?
    → Tag v999.999.999-typo tidak ada di Docker Hub
  
  Why 3: Kenapa tag salah sampai production?
    → Typo di manifest YAML (deployment script)
  
  Why 4: Kenapa typo tidak terdeteksi?
    → Tidak ada verifikasi image existence di CI
  
  Why 5: Kenapa CI tidak verifikasi?
    → CI hanya jalankan unit test, tidak ada image registry check

🎯 ROOT CAUSE: Tag typo tidak terdeteksi sebelum deploy
```

### Tabel Diagnosis ImagePullBackOff

| Event Message Fragment | Root Cause | Solusi |
|---|---|---|
| `failed to resolve reference` + `404 Not Found` | Tag tidak ada | Fix tag ke yang valid |
| `pull access denied` + `may require docker login` | Private registry, no Secret | Buat docker-registry Secret + imagePullSecrets |
| `dial tcp: lookup registry.xx: no such host` | DNS registry tidak resolve | Fix DNS atau pakai IP |
| `i/o timeout` saat pull | Network/firewall block | Buka firewall ke registry |
| `toomanyrequests` | Rate limit Docker Hub anonymous | Login dengan Docker Hub account atau pakai GHCR |
| `no space left on device` | Disk node penuh untuk image cache | Cleanup image, tambah disk |

---

## 🛠️ Mitigation

### Opsi A — Fix Image Tag (Paling Umum)

```bash
# Patch deployment dengan tag yang benar
kubectl set image deployment/imagepull-app -n insiden-lab   app=nginx:1.27-alpine

# Verify
kubectl set image deployment/imagepull-app -n insiden-lab app --list
# Output: app=nginx:1.27-alpine

# Tunggu pull selesai
kubectl get pods -n insiden-lab -l app=imagepull-app -w
```

### Opsi B — Buat docker-registry Secret untuk Private Registry

```bash
# 1. Buat Secret dari credentials (pakai docker login)
kubectl create secret docker-registry regcred -n insiden-lab   --docker-server=ghcr.io   --docker-username=andi   --docker-password=<personal-access-token>   --docker-email=andi@perusahaan.com

# 2. Patch deployment untuk pakai secret
kubectl patch deployment/private-app -n insiden-lab --type=json   -p='[{"op":"add","path":"/spec/template/spec/imagePullSecrets","value":[{"name":"regcred"}]}]'

# Verify
kubectl get pod -n insiden-lab -l app=private-app -w
```

### Opsi C — Buat Secret dari Docker Config JSON (untuk CI/CD)

```bash
# Generate docker config dari login existing
docker login ghcr.io -u andi -p <token>
# File ~/.docker/config.json akan tercipta

# Buat Secret dari file
kubectl create secret generic regcred -n insiden-lab   --from-file=.dockerconfigjson=$HOME/.docker/config.json   --type=kubernetes.io/dockerconfigjson
```

### Opsi D — Patch via Patch File (lebih bersih)

```bash
# Buat patch file
cat > /tmp/patch.yaml <<EOF
spec:
  template:
    spec:
      imagePullSecrets:
      - name: regcred
EOF

# Apply patch
kubectl patch deployment/private-app -n insiden-lab   --patch-file /tmp/patch.yaml
```

### Opsi E — Login di Semua Node (untuk Image Already Cached Strategy)

```bash
# Login di setiap node cluster
ssh node1 "docker login ghcr.io -u andi -p <token>"
ssh node2 "docker login ghcr.io -u andi -p <token>"
# Setelah login, pull manual:
ssh node1 "docker pull ghcr.io/perusahaan-rahasia/secret-app:v1.0.0"
```

---

## 🛡️ Prevention

### Prevention 1 — Image Tag Verification di CI

```yaml
# .gitlab-ci.yml
validate-image:
  stage: validate
  image: docker:24
  services:
  - docker:24-dind
  script:
  - |
    # Extract image reference dari manifests
    IMAGE=$(yq eval '.spec.template.spec.containers[0].image' k8s/deployment.yaml)
    echo "Verifying image: $IMAGE"
    
    # Cek image exists di registry
    if docker manifest inspect "$IMAGE" > /dev/null 2>&1; then
      echo "✅ Image exists"
    else
      echo "❌ Image NOT FOUND: $IMAGE"
      exit 1
    fi
  allow_failure: false
```

### Prevention 2 — Pakai Image Digest (Immutable) Daripada Tag

```yaml
spec:
  containers:
  - name: app
    # SHA256 digest SELALU valid (kalau image ada)
    image: nginx@sha256:4f8b9f6f9b7c8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b2a1f0e9d8c7b6a5f4e3d
```

**Cara dapat digest:**
```bash
docker inspect --format='{{index .RepoDigests 0}}' nginx:1.27-alpine
# Output: nginx@sha256:abc123...
```

### Prevention 3 — ServiceAccount dengan imagePullSecrets Otomatis

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: insiden-lab
imagePullSecrets:
- name: regcred
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-app
  namespace: insiden-lab
spec:
  template:
    spec:
      serviceAccountName: app-sa    # ← otomatis pakai Secret dari SA
```

**Keuntungan:** Semua pod yang pakai SA ini otomatis punya kredensial. Tidak perlu set `imagePullSecrets` per-pod.

### Prevention 4 — Image Scanning + Vulnerability Check

```yaml
# GitLab CI
trivy-scan:
  stage: security
  image: aquasec/trivy:latest
  script:
  - trivy image --exit-code 1 --severity HIGH,CRITICAL $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  allow_failure: false
```

### Prevention 5 — Local Registry Mirror di Cluster

```yaml
# /etc/rancher/k3s/registries.yaml di setiap node
mirrors:
  docker.io:
    endpoint:
      - "https://registry-mirror.internal:5000"
  ghcr.io:
    endpoint:
      - "https://ghcr-mirror.internal:5000"
configs:
  "registry-mirror.internal:5000":
    tls:
      insecure_skip_verify: true
```

---

## 🧪 Verifikasi Recovery

```bash
# 1. Pod Running
kubectl get pods -n insiden-lab -l app=imagepull-app
# NAME                            READY   STATUS    RESTARTS   AGE
# imagepull-app-7f9c8d9f8-abcd1   1/1     Running   0          30s

# 2. Image sudah ter-pull di node
sudo crictl images | grep nginx
# Output: docker.io/library/nginx   1.27-alpine   ...

# 3. Alert resolved
# Cek Slack: [RESOLVED] ImagePullBackOff

# 4. Test pull manual untuk verifikasi Secret bekerja
kubectl run test-pull -n insiden-lab --rm -it   --image=ghcr.io/perusahaan-rahasia/secret-app:v1.0.0   --command -- echo "pull OK"
# Output: pull OK
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/04-imagepull-bad-tag.yaml
kubectl delete -f manifests/04b-imagepull-private-no-secret.yaml
kubectl delete secret regcred -n insiden-lab
```

---

## 📖 Rangkuman

| Aspek | Catatan |
|---|---|
| **Symptom utama** | STATUS=ErrImagePull / ImagePullBackOff |
| **Cara diagnosis** | `kubectl describe pod` → Events → Failed pulling |
| **Penyebab umum** | Tag typo, no credentials, registry down, rate limit |
| **Mitigation** | Fix tag, create docker-registry Secret |
| **Prevention terbaik** | CI image verification + digest pinning + ServiceAccount |
| **Alert yang firing** | `ImagePullBackOff` (Minggu 8) ✅ |

---

## ➡️ Modul 06: FailedMount

Sekarang kita masuk incident terakhir minggu ini: **FailedMount** — pod sudah punya container, tapi **gagal mount volume/PVC** sehingga container tidak bisa start.

👉 Lanjut ke `06-incident-failedmount.md`
