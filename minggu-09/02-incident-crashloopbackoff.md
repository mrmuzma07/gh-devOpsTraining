# Minggu 9 — Modul 02: Incident CrashLoopBackOff

> **"Aplikasi tidak jalan sama sekali, restart terus-menerus!"**

Ini adalah incident paling **klasik dan paling sering** dijumpai. CrashLoopBackOff artinya K8s mencoba menjalankan container, container-nya langsung mati (exit dengan error), lalu K8s tunggu, coba lagi, tunggu lebih lama, coba lagi, dan seterusnya dengan pola **exponential backoff**.

Bayangkan seperti tombol power yang kamu tekan, lampunya nyala 1 detik lalu mati, kamu tekan lagi, nyala 1 detik lalu mati lagi, dst. Itu CrashLoopBackOff.

---

## 🎯 Tujuan Modul

1. Mengenali **symptom** CrashLoopBackOff dari output `kubectl`
2. Melakukan **investigation** terstruktur dengan `describe`, `logs --previous`, dan `events`
3. Membedakan **5 penyebab umum** CrashLoopBackOff
4. Menentukan **mitigation** yang tepat untuk tiap penyebab
5. Menerapkan **prevention** berupa startupProbe, validasi CI, dan proper config management

---

## 📦 Simulasi: Reproduce Incident

Kita akan deploy sebuah aplikasi yang sengaja dibuat crash untuk latihan.

### File: `manifests/01-crashloop-bad-config.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: insiden-lab
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: crashloop-app
  namespace: insiden-lab
  labels:
    app: crashloop-app
    lab: insiden-1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: crashloop-app
  template:
    metadata:
      labels:
        app: crashloop-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "echo Fatal missing DATABASE_URL; exit 1"]
        env:
        - name: DATABASE_URL
          value: ""
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
        livenessProbe:
          exec:
            command: ["sh", "-c", "test -n DATABASE_URL"]
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Deploy & Observe

```bash
kubectl apply -f manifests/01-crashloop-bad-config.yaml

kubectl get pods -n insiden-lab
```

---

## 🔍 Symptoms

- READY 0/1, STATUS CrashLoopBackOff, RESTARTS naik
- Alert PodCrashLoop firing
- Endpoints kosong (no ready pod)

---

## 🕵️ Investigation

### Step 1 — Nama pod
```bash
kubectl get pods -n insiden-lab -l app=crashloop-app
```

### Step 2 — describe
```bash
kubectl describe pod crashloop-app-xxx -n insiden-lab
```
Lihat bagian Events: Back-off restarting failed container

### Step 3 — logs --previous
```bash
kubectl logs crashloop-app-xxx -n insiden-lab --previous
```
Output: Fatal: missing required env var DATABASE_URL

### Step 4 — events
```bash
kubectl get events -n insiden-lab --sort-by=.lastTimestamp
```

### Step 5 — Cross-check Loki
```logql
{namespace="insiden-lab"} |= "DATABASE_URL"
```

---

## 🎯 Root Cause

5 Whys:
1. Kenapa CrashLoopBackOff? Container exit 1
2. Kenapa exit 1? Validasi DATABASE_URL gagal
3. Kenapa DATABASE_URL kosong? Manifest value sengaja kosong
4. Kenapa kosong sampai production? CI/CD tidak inject
5. Kenapa CI/CD tidak inject? Tidak ada validasi required env

ROOT CAUSE: Missing env var injection di CI/CD pipeline.

### 5 Penyebab Umum

1. Missing env var / ConfigMap / Secret
2. Bug code (panic, exception)
3. Liveness probe gagal terus
4. Image command/args salah
5. Database/dependency tidak ready

---

## 🛠️ Mitigation

### Opsi A — Fix env var
```bash
kubectl set env deployment/crashloop-app -n insiden-lab   DATABASE_URL="postgres://user:pass@postgres:5432/db"
```

### Opsi B — Rollback
```bash
kubectl rollout undo deployment/crashloop-app -n insiden-lab
```

### Opsi C — Pause untuk Investigasi
```bash
kubectl scale deployment/crashloop-app -n insiden-lab --replicas=0
# fix, lalu:
kubectl scale deployment/crashloop-app -n insiden-lab --replicas=2
```

### Opsi D — Delete pod
```bash
kubectl delete pod crashloop-app-xxx -n insiden-lab
```

---

## 🛡️ Prevention

### 1. startupProbe
```yaml
startupProbe:
  exec:
    command: ["sh", "-c", "test -n DATABASE_URL"]
  periodSeconds: 5
  failureThreshold: 12
```

### 2. Validasi env var di CI
```bash
for var in DATABASE_URL REDIS_URL JWT_SECRET; do
  [ -z "" ] && { echo "ERROR:  missing"; exit 1; }
done
```

### 3. initContainer check
```yaml
initContainers:
- name: check-config
  image: busybox:1.36
  command: ["sh", "-c", "[ -n DATABASE_URL ]"]
```

### 4. ArgoCD PreSync hook

### 5. Dokumentasi required env

---

## 🧪 Verifikasi Recovery

```bash
kubectl get pods -n insiden-lab
kubectl logs -n insiden-lab -l app=crashloop-app --tail=5
kubectl get endpoints crashloop-app -n insiden-lab
```

---

## 🧹 Cleanup

```bash
kubectl delete -f manifests/01-crashloop-bad-config.yaml
```

---

## 📖 Rangkuman

- Symptom: CrashLoopBackOff, RESTARTS naik
- Tool: describe, logs --previous
- Root cause umum: missing env var, bug code
- Prevention: startupProbe, CI validation, initContainer

---

## ➡️ Modul 03: OOMKilled
