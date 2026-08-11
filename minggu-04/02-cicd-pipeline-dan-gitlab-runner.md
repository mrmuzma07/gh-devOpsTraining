# Modul 02: CI/CD Pipeline & GitLab Integration

> **Target Pembelajaran:** Memahami pemisahan peran antara **Continuous Integration (CI)** dan **Continuous Deployment (CD)** pada GitOps, serta konfigurasi file `.gitlab-ci.yml`.

---

## 1. Pemisahan Peran: CI vs CD dalam GitOps

Dalam arsitektur GitOps modern, peran **GitLab CI** dan **ArgoCD** dibagi secara tegas:

```mermaid
graph LR
    subgraph CI_Pipeline [GitLab CI/CD - Integration]
        Code[Source Code Go] --> Test[Unit Test]
        Test --> Build[Build OCI Image via Podman/Kaniko]
        Build --> Push[Push Image ke Container Registry]
        Push --> UpdateGit[Update Image Tag di GitOps Repo]
    end

    subgraph CD_Pipeline [ArgoCD - Deployment]
        UpdateGit --> ArgoDetect[ArgoCD Detect Changes]
        ArgoDetect --> Sync[Auto-Sync to k3s Cluster]
    end
```

### Tugas Masing-Masing Komponen:
- **GitLab CI (Continuous Integration):** FOKUS pada kode program (testing, linting, pembentukan Docker image, dan update versi tag image).
- **ArgoCD (Continuous Deployment):** FOKUS pada status cluster (memastikan manifest Kubernetes di cluster selalu cocok dengan isi Git Repository).

---

## 2. Praktik Terbaik: Pemisahan Repository

Sangat disarankan untuk memisahkan repository menjadi 2 bagian:

1. **Application Repository (`app-go`):**
   - Berisi kode sumber aplikasi Go, unit test, dan file `.gitlab-ci.yml`.
   - Diakses oleh pengembang aplikasi (*Developers*).

2. **GitOps Configuration Repository (`infra-gitops`):**
   - Berisi Helm Charts, values file per-environment (`values-dev.yaml`, `values-prod.yaml`), dan file manifest ArgoCD.
   - Diakses oleh tim DevOps/SRE dan di-poll oleh ArgoCD.

---

## 3. Anatomi File Pipeline `.gitlab-ci.yml`

Berikut adalah contoh alur pipeline `.gitlab-ci.yml` yang bertugas mem-build image dan memperbarui versi tag di repository GitOps:

```yaml
stages:
  - test
  - build
  - update-manifest

variables:
  CONTAINER_IMAGE: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA

# 1. Stage Testing
unit-test:
  stage: test
  image: golang:1.22-alpine
  script:
    - go test -v ./...

# 2. Stage Build & Push OCI Image
build-image:
  stage: build
  image: quay.io/podman/stable
  script:
    - podman login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - podman build -t $CONTAINER_IMAGE -f minggu-02/app/Dockerfile .
    - podman push $CONTAINER_IMAGE

# 3. Stage Update GitOps Manifest Repository
trigger-gitops:
  stage: update-manifest
  image: alpine/git
  script:
    - git clone https://oauth2:$GITOPS_TOKEN@gitlab.com/company/infra-gitops.git
    - cd infra-gitops
    # Update tag image di values-dev.yaml secara otomatis
    - sed -i "s/tag:.*/tag:\ \"$CI_COMMIT_SHORT_SHA\"/" clusters/dev/values.yaml
    - git config user.name "GitLab CI Bot"
    - git config user.email "ci-bot@company.com"
    - git commit -am "chore(deps): bump go-app image tag to $CI_COMMIT_SHORT_SHA"
    - git push origin main
```

---

## Ringkasan Modul 02

- **GitLab CI** bertanggung jawab menguji kode, mem-build container image, dan meng-commit versi image baru ke Git.
- **Pemisahan Repository** (*App Repo vs Config Repo*) menjaga keamanan dan kebersihan histori commit.
- **ArgoCD** secara otomatis mendeteksi commit baru di Config Repo lalu meng-apply perubahan ke cluster tanpa campur tangan CI Runner.
