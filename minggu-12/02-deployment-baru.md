# Modul 02 — Deployment Baru: GitOps Workflow End-to-End

> **Satu kalimat:** Deployment baru bukan `kubectl apply` manual, melainkan **rantai otomatis**: `git push → CI Pipeline → Image Registry → ArgoCD Sync → Kubernetes Update → Health Check Verification`.

Di minggu-minggu sebelumnya kita sudah setup setiap komponen secara terpisah. Sekarang di Week 12, kita lihat **keseluruhan flow** yang terjadi saat developer push kode baru. Anda akan memahami setiap **stage**, **gate**, dan **verification step** sehingga deployment terasa **boring and predictable** (sesuai motto SRE).

---

## 🎯 Learning Outcomes

1. Memahami **GitOps workflow** lengkap: dari `git push` sampai pod baru running.
2. Mendesain **GitLab CI pipeline** untuk build, test, dan security scan.
3. Mengkonfigurasi **ArgoCD Application** untuk auto-sync.
4. Mengimplementasikan **canary deployment** strategy (opsional, advanced).
5. Melakukan **post-deployment verification** otomatis (smoke test + SLO check).

---

## 1. 🛤️ The Golden Path: 7 Tahap Deployment

```
┌──────────────────────────────────────────────────────────────────┐
│              THE GOLDEN PATH OF DEPLOYMENT                        │
│                                                                  │
│  Tahap 1:  git push / merge request ke main branch              │
│       ↓                                                          │
│  Tahap 2:  GitLab CI trigger (lint → test → build → scan)        │
│       ↓                                                          │
│  Tahap 3:  Docker image di-build & push ke Registry              │
│       ↓                                                          │
│  Tahap 4:  Update manifest di Git repo (image tag baru)          │
│       ↓                                                          │
│  Tahap 5:  ArgoCD detect perubahan, sync ke cluster              │
│       ↓                                                          │
│  Tahap 6:  Kubernetes rolling update, health check pass          │
│       ↓                                                          │
│  Tahap 7:  Post-deployment verification (smoke test + SLO)       │
│                                                                  │
│  Total waktu ideal: 8-12 menit (auto)                            │
└──────────────────────────────────────────────────────────────────┘
```

### Visualisasi Lebih Detail

```mermaid
sequenceDiagram
    actor Dev as 👨💻 Developer
    participant GL as GitLab Repo
    participant CI as GitLab CI
    participant Reg as Container Registry
    participant Argo as ArgoCD
    participant K8s as Kubernetes
    participant Graf as Grafana

    Dev->>GL: 1. Push code (git push origin main)
    GL->>CI: 2. Trigger pipeline webhook
    CI->>CI: 3. Lint (golangci-lint)
    CI->>CI: 4. Test (go test ./...)
    CI->>CI: 5. Build (docker build)
    CI->>Reg: 6. Push image (yourname/app:v1.2.3)
    CI->>GL: 7. Update manifests with new tag
    Argo->>GL: 8. Poll repo every 3 min (or webhook)
    Argo->>K8s: 9. kubectl apply (rolling update)
    K8s->>K8s: 10. Spin up new pods (readiness probe)
    K8s-->>Argo: 11. Health check pass
    Argo-->>Dev: 12. Notification: Sync Successful
    Dev->>Graf: 13. Verify SLO (latency, error rate)
```

---

## 2. 📁 Repository Structure

Repository mini-prod-platform Anda harus punya struktur ini:

```text
mini-prod-platform/
├── .gitlab-ci.yml                 # CI pipeline definition
├── README.md
├── app/                           # Go application source
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── handlers/
│   ├── middleware/
│   └── tests/
├── charts/                        # Helm chart (Week 3)
│   └── checkout-service/
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── values-prod.yaml
│       └── templates/
│           ├── deployment.yaml
│           ├── service.yaml
│           ├── configmap.yaml
│           └── ingress.yaml
├── manifests/                     # Kustomize / ArgoCD source
│   ├── base/
│   │   ├── kustomization.yaml
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── overlays/
│       ├── dev/
│       │   └── kustomization.yaml
│       ├── staging/
│       │   └── kustomization.yaml
│       └── prod/
│           └── kustomization.yaml
└── argocd/
    └── applications/
        ├── checkout-service.yaml
        └── postgres.yaml
```

---

## 3. 🛠️ GitLab CI Pipeline (`.gitlab-ci.yml`)

File: `minggu-12/manifests/02-gitlab-ci.yaml`

```yaml
# .gitlab-ci.yml — Multi-stage pipeline for Go API
stages:
  - lint
  - test
  - build
  - security
  - update-manifest
  - notify

variables:
  IMAGE_NAME: registry.gitlab.local/mini-prod/checkout-service
  IMAGE_TAG: ${CI_COMMIT_SHORT_SHA}
  HELM_CHART_PATH: ./charts/checkout-service

# === CACHE: Go modules ===
cache:
  key: ${CI_COMMIT_REF_SLUG}
  paths:
    - .cache/go-build/
    - .cache/gocache/
    - app/vendor/

# ===========================================
# STAGE 1: LINT
# ===========================================
lint:
  stage: lint
  image: golang:1.22-alpine
  before_script:
    - cd app
    - go install github.com/golangci/golangci-lint/cmd/[EMAIL_REDACTED]@latest
  script:
    - golangci-lint run --timeout 5m ./...
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

# ===========================================
# STAGE 2: TEST
# ===========================================
unit-test:
  stage: test
  image: golang:1.22-alpine
  services:
    - postgres:15-alpine
  variables:
    POSTGRES_DB: test_db
    POSTGRES_USER: test
    POSTGRES_PASSWORD: test
  before_script:
    - cd app
    - apk add --no-cache git
  script:
    - go mod download
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out | tail -1
  coverage: '/total:\s+\(statements\)\s+(\d+.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: app/coverage.xml
    paths:
      - app/coverage.out
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

helm-lint:
  stage: test
  image: alpine/helm:3.14
  script:
    - helm lint ${HELM_CHART_PATH}
    - helm template ${HELM_CHART_PATH} > /tmp/manifest.yaml
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

# ===========================================
# STAGE 3: BUILD
# ===========================================
build-image:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  variables:
    DOCKER_TLS_CERTDIR: "/certs"
  before_script:
    - echo "$REGISTRY_PASSWORD" | docker login registry.gitlab.local -u $REGISTRY_USER --password-stdin
  script:
    - cd app
    - docker build -t ${IMAGE_NAME}:${IMAGE_TAG} -t ${IMAGE_NAME}:latest .
    - docker push ${IMAGE_NAME}:${IMAGE_TAG}
    - docker push ${IMAGE_NAME}:latest
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

# ===========================================
# STAGE 4: SECURITY SCAN
# ===========================================
trivy-scan:
  stage: security
  image: aquasec/trivy:0.48
  script:
    - trivy image --severity HIGH,CRITICAL --exit-code 1 ${IMAGE_NAME}:${IMAGE_TAG}
  allow_failure: true
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

# ===========================================
# STAGE 5: UPDATE MANIFEST (TRIGGER ARGOCD)
# ===========================================
update-manifest:
  stage: update-manifest
  image: alpine:3.19
  before_script:
    - apk add --no-cache git
    - git config --global user.email "[email protected]"
    - git config --global user.name "GitLab CI"
  script:
    - git clone https://gitlab-ci-token:${CI_JOB_TOKEN}@gitlab.local/mini-prod/manifests.git /tmp/manifests
    - cd /tmp/manifests
    # Update image tag di kustomization.yaml
    - sed -i "s|newTag: .*|newTag: ${IMAGE_TAG}|" overlays/prod/kustomization.yaml
    - git add overlays/prod/kustomization.yaml
    - git commit -m "ci: bump checkout-service to ${IMAGE_TAG}"
    - git push origin main
  rules:
    - if: $CI_COMMIT_BRANCH == "main"

# ===========================================
# STAGE 6: NOTIFY
# ===========================================
notify-success:
  stage: notify
  image: alpine:3.19
  script:
    - apk add --no-cache curl
    - |
      curl -X POST ${SLACK_WEBHOOK_URL} \
        -H 'Content-Type: application/json' \
        -d "{
          \"text\": \"✅ Deployment ${IMAGE_TAG} succeeded\",
          \"channel\": \"#deployments\"
        }"
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - when: on_failure
```

---

## 4. 🐳 Dockerfile Multi-Stage (Best Practice)

File: `minggu-12/manifests/02-Dockerfile`

```dockerfile
# === Stage 1: Build ===
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Copy go.mod dulu untuk caching layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main .

# === Stage 2: Runtime ===
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/templates /app/templates

USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/main"]
```

**Image size:** ~25MB (vs 800MB jika pakai `golang:1.22` langsung).

---

## 5. 🚀 ArgoCD Application Definition

File: `minggu-12/manifests/02-argocd-app.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: checkout-service
  namespace: argocd
  labels:
    app: checkout-service
    environment: prod
spec:
  project: default
  
  source:
    repoURL: https://gitlab.local/mini-prod/manifests.git
    targetRevision: main
    path: overlays/prod
    
    # Plugin Helm support
    plugin:
      name: kustomize-build-with-helm
    
    # OR menggunakan Helm directly:
    # chart:
    #   spec:
    #     chart: checkout-service
    #     version: 1.2.3
    #     sourceRef:
    #       repoURL: https://gitlab.local/mini-prod/charts.git
    #       chart: checkout-service
  
  destination:
    server: https://kubernetes.default.svc
    namespace: prod
  
  syncPolicy:
    automated:
      prune: true          # Hapus resource yang sudah tidak ada di git
      selfHeal: true       # Auto-sync jika manual changes terdeteksi
      allowEmpty: false
    
    syncOptions:
      - CreateNamespace=true
      - PrunePropagationPolicy=foreground
      - ApplyOutOfSyncOnly=true
    
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  
  revisionHistoryLimit: 10
```

### 5.1 Sync Wave (Urutan Deployment)

Untuk menghindari race condition (mis. API deploy sebelum Postgres ready), gunakan **sync waves**:

```yaml
# Di Helm values, tambahkan annotation:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkout-service
  annotations:
    argocd.argoproj.io/sync-wave: "2"  # Deploy setelah wave 0 dan 1
```

```text
Sync Wave 0: Namespace, ServiceAccount, Secret
Sync Wave 1: ConfigMap, PVC, StatefulSet (postgres)
Sync Wave 2: Deployment (Go API) - depend on Postgres
Sync Wave 3: HPA, VPA, NetworkPolicy
```

---

## 6. 🧪 Post-Deployment Verification

Setelah ArgoCD sync selesai, **JANGAN langsung dianggap selesai**. Lakukan verifikasi:

File: `minggu-12/manifests/02-post-deploy-verify.sh`

```bash
#!/usr/bin/env bash
set -uo pipefail

# Post-Deployment Verification
# Dipanggil otomatis oleh CI setelah deployment, atau manual oleh engineer

SERVICE=${1:-checkout-service}
NAMESPACE=${2:-prod}
EXPECTED_REPLICAS=${3:-3}
MAX_WAIT_SECONDS=180

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0

check() {
    if [ "$1" -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
        PASS=$((PASS+1))
    else
        echo -e "${RED}✗${NC} $2"
        FAIL=$((FAIL+1))
    fi
}

echo "==========================================="
echo "  Post-Deployment Verification: $SERVICE"
echo "==========================================="

# === 1. Wait for rollout ===
echo "[1/6] Waiting for rollout..."
kubectl rollout status deployment/$SERVICE -n $NAMESPACE --timeout=${MAX_WAIT_SECONDS}s
check $? "Rollout completed within ${MAX_WAIT_SECONDS}s"

# === 2. Check replica count ===
echo ""
echo "[2/6] Checking replica count..."
READY=$(kubectl get deployment $SERVICE -n $NAMESPACE -o jsonpath='{.status.readyReplicas}')
DESIRED=$(kubectl get deployment $SERVICE -n $NAMESPACE -o jsonpath='{.spec.replicas}')
[ "$READY" = "$DESIRED" ] && check 0 "Replicas ready: $READY/$DESIRED" || check 1 "Replicas: $READY/$DESIRED"

# === 3. Check no CrashLoopBackOff ===
echo ""
echo "[3/6] Checking for CrashLoopBackOff..."
CRASHING=$(kubectl get pods -n $NAMESPACE -l app=$SERVICE --no-headers | grep -c -E "CrashLoopBackOff|Error" || true)
[ "$CRASHING" = "0" ] && check 0 "No CrashLoopBackOff pods" || check 1 "$CRASHING pods in CrashLoopBackOff"

# === 4. Health check ===
echo ""
echo "[4/6] Health check..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://$SERVICE.$NAMESPACE.svc.cluster.local:8080/healthz)
[ "$HTTP_STATUS" = "200" ] && check 0 "Healthz returns 200" || check 1 "Healthz returned $HTTP_STATUS"

# === 5. Functional smoke test ===
echo ""
echo "[5/6] Functional smoke test..."
# Test root endpoint
ROOT_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://$SERVICE.$NAMESPACE.svc.cluster.local:8080/)
[ "$ROOT_STATUS" = "200" ] && check 0 "Root endpoint OK" || check 1 "Root endpoint $ROOT_STATUS"

# Test API endpoint
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://$SERVICE.$NAMESPACE.svc.cluster.local:8080/api/v1/products)
[ "$API_STATUS" = "200" ] && check 0 "Products API OK" || check 1 "Products API $API_STATUS"

# === 6. SLO check (wait 2 minutes for metrics) ===
echo ""
echo "[6/6] SLO check (2 minute wait for fresh metrics)..."
echo "  Waiting 2 minutes for fresh metrics to accumulate..."
sleep 120

# Check error rate via Prometheus
PROM_URL="http://prometheus-operated.monitoring.svc:9090"
ERROR_RATE=$(curl -sG $PROM_URL/api/v1/query \
    --data-urlencode "query=sum(rate(http_requests_total{job=\"$SERVICE\",code=~\"5xx\"}[5m])) / sum(rate(http_requests_total{job=\"$SERVICE\"}[5m]))" \
    | jq -r '.data.result[0].value[1] // "0"')

# Compare to SLO 0.1% (0.001)
SLO_THRESHOLD=0.001
IS_OK=$(echo "$ERROR_RATE < $SLO_THRESHOLD" | bc)
[ "$IS_OK" = "1" ] && check 0 "Error rate $ERROR_RATE < SLO 0.001" || check 1 "Error rate $ERROR_RATE exceeds SLO 0.001"

# Check p95 latency
P95_LATENCY=$(curl -sG $PROM_URL/api/v1/query \
    --data-urlencode "query=histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job=\"$SERVICE\"}[5m])) by (le))" \
    | jq -r '.data.result[0].value[1] // "0"')
P95_MS=$(echo "$P95_LATENCY * 1000" | bc -l)
LATENCY_OK=$(echo "$P95_MS < 200" | bc -l)
[ "$LATENCY_OK" = "1" ] && check 0 "p95 latency ${P95_MS}ms < SLO 200ms" || check 1 "p95 latency ${P95_MS}ms exceeds SLO 200ms"

# === SUMMARY ===
echo ""
echo "==========================================="
echo "  PASS: $PASS  FAIL: $FAIL"
echo "==========================================="

[ "$FAIL" -eq 0 ] && exit 0 || exit 1
```

### 6.1 Auto-rollback jika Verification Gagal

Tambahkan step di GitLab CI setelah deployment:

```yaml
post-deploy-verify:
  stage: notify
  image: alpine:3.19
  before_script:
    - apk add --no-cache bash curl jq bc
  script:
    - ./minggu-12/manifests/02-post-deploy-verify.sh checkout-service prod 3
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
```

---

## 7. 🎯 Deployment Strategies

### 7.1 Recreate vs Rolling Update vs Canary

| Strategy | Downtime | Risk | Cocok untuk |
|---|---|---|---|
| **Recreate** | Ya (ada jeda) | Medium | Dev environment |
| **Rolling Update** (default) | Zero | Medium | Production standard |
| **Blue/Green** | Zero | Low (rollback instan) | Critical services |
| **Canary** | Zero | Lowest (incremental) | Risky changes |
| **A/B Testing** | Zero | - | Feature flag driven |

### 7.2 Konfigurasi Rolling Update

```yaml
# Helm values
deployment:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1          # Maksimal 1 pod extra saat rollout
      maxUnavailable: 0    # TIDAK ADA pod yang boleh unavailable
```

### 7.3 Canary dengan Argo Rollouts (Advanced)

File: `minggu-12/manifests/02-argorollout-canary.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: checkout-service
  namespace: prod
spec:
  replicas: 5
  strategy:
    canary:
      steps:
        - setWeight: 10      # 10% traffic ke canary
        - pause: { duration: 5m }   # Tunggu 5 menit observasi
        - setWeight: 25
        - pause: { duration: 5m }
        - setWeight: 50
        - pause: { duration: 5m }
        - setWeight: 100
  selector:
    matchLabels:
      app: checkout-service
  template:
    metadata:
      labels:
        app: checkout-service
    spec:
      containers:
      - name: app
        image: registry.gitlab.local/mini-prod/checkout-service:1.2.3
        # ... rest of container spec
```

**Proses canary:**
1. Deploy v1.2.3 (canary) alongside v1.2.2 (stable).
2. Alihkan 10% traffic ke canary.
3. Monitor SLO (error rate, latency) selama 5 menit.
4. Jika OK, naikkan ke 25%, 50%, 100%.
5. Jika ada masalah di salah satu step, **auto-rollback**.

---

## 8. 📋 Cheat Sheet: Deployment Flow

```text
Tahap    | Siapa           | Apa yang terjadi                  | Waktu
---------|-----------------|-----------------------------------|--------
1. Code  | Developer       | git push ke GitLab                | -
2. CI    | GitLab Runner   | lint → test → build → scan        | 3-5 min
3. Image | Registry        | Image tag: abc1234                | -
4. Sync  | ArgoCD          | Detect image tag baru, sync       | 30s-2 min
5. Roll  | Kubernetes      | Rolling update pods               | 1-3 min
6. Verify| CI + Prometheus | Smoke test + SLO check            | 2-3 min
Total    |                 |                                   | ~10 min
```

---

## 9. ✏️ Latihan Mandiri

1. **Setup GitLab CI pipeline** untuk servis Anda. Pastikan setiap stage punya `rules` yang benar.
2. **Buat Dockerfile multi-stage** untuk Go API, ukuran image harus < 50MB.
3. **Konfigurasi ArgoCD** dengan `syncPolicy.automated: true` dan `selfHeal: true`.
4. **Implementasikan post-deploy verification** dan test trigger rollback otomatis.
5. **Buat Argo Rollout** dengan canary strategy 10→50→100% untuk production.

---

**Lanjut ke Modul 03:** [03-load-test.md](./03-load-test.md) — K6 load test comprehensive untuk validasi SLO dan identifikasi bottleneck.
