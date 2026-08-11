# Modul 05 — Postmortem: Belajar dari Kegagalan Tanpa Menyalahkan

> **Satu kalimat:** **Postmortem** adalah ritual **blameless** setelah insiden yang tujuannya bukan mencari siapa yang salah, tapi **memahami mengapa sistem gagal** sehingga kita bisa mencegah insiden serupa di masa depan.

Filosofi inti postmortem: **manusia tidak gagal, sistem yang gagal**. Seorang on-call engineer yang membuat keputusan salah pukul 3 pagi BUKAN orang yang salah — yang salah adalah sistem yang membuatnya harus memutuskan dalam kondisi panik tanpa informasi yang cukup. Peran postmortem adalah memperbaiki **sistem** (process, tooling, runbook, alert), bukan menghukum **manusia**.

---

## 🎯 Learning Outcomes

1. Memahami filosofi **blameless postmortem** ala Google SRE.
2. Menulis **timeline insiden** yang akurat dari sumber observabilitas (Mimir, Loki, Tempo, kubectl).
3. Melakukan **5-Whys analysis** untuk menemukan root cause sistemik.
4. Mendesain **action items** yang SMART dan ditugaskan ke owner jelas.
5. Menggunakan template postmortem Google SRE untuk dokumentasi konsisten.

---

## 1. 🧠 Filosofi Blameless Postmortem

### 1.1 Konteks Historis

Postmortem di industri aviation (NASA) sudah ada sejak 1970-an. NASA menemukan bahwa **95% kecelakaan pesawat** bukan karena pilot "bodoh", tapi karena:
- Sistem yang mengharuskan pilot membuat split-second decision
- Tooling yang tidak menampilkan informasi yang tepat
- Proses yang membuat manusia kelelahan
- Interface yang counterintuitive

Prinsip ini diadopsi oleh Google SRE dan jadi standar industri software.

### 1.2 Aturan Utama Blameless

```text
┌──────────────────────────────────────────────────────────────────┐
│ ATURAN BLAMELESS POSTMORTEM                                      │
│                                                                  │
│ 1. Fokus pada KONTRIBUSI SISTEM, bukan niat individu             │
│    ❌ "Andi salah hapus database"                                 │
│    ✅ "Tidak ada backup otomatis + console akses terlalu permisif"│
│                                                                  │
│ 2. Asumsikan setiap orang bertindak dengan niat baik              │
│    ❌ "Kok gak teliti?"                                           │
│    ✅ "Apa yang membuat orang ini percaya keputusannya benar?"   │
│                                                                  │
│ 3. Cari MULTIPLE contributing factors                            │
│    ❌ "Penyebab: human error"                                    │
│    ✅ "5 faktor: tool, training, alert, runbook, fatigue"         │
│                                                                  │
│ 4. Solusi harus MENCEGAH, bukan MENYEMBUHKAN                     │
│    ❌ "Andi harus lebih hati-hati"                                │
│    ✅ "Tambah 2-eye review + automated test + alert pre-merge"   │
│                                                                  │
│ 5. Dokumentasi dipublikasi ke SELURUH tim                        │
│    ❌ Disembunyikan untuk jaga reputasi                          │
│    ✅ Dibagikan agar semua belajar                                │
└──────────────────────────────────────────────────────────────────┘
```

### 1.3 Anatomi Postmortem Google SRE

Template resmi Google SRE Book:

```text
1. Summary (ringkasan 1 paragraf)
2. Impact (dampak kuantitatif: user affected, downtime, revenue loss)
3. Timeline (kronologis dengan timestamp)
4. Root Cause (penyebab utama + contributing factors)
5. Trigger (apa yang memulai insiden)
6. Resolution (apa yang memperbaiki)
7. Detection (bagaimana kita tahu ada masalah)
8. Response (apa yang dilakukan)
9. Recovery (kapan kembali normal)
10. Lessons Learned (apa yang kita pelajari)
11. Action Items (tindakan preventif dengan owner)
```

---

## 2. 📋 Template Postmortem (Siap Pakai)

File: `minggu-11/templates/postmortem-template.md`

```markdown
# Postmortem: [Nama Insiden Singkat]

| Metadata | Value |
|---|---|
| **Postmortem ID** | PM-2026-08-XXX |
| **Date of Incident** | 2026-08-10 |
| **Author(s)** | @sre-oncall |
| **Status** | Draft / In Review / Action Items In Progress / Closed |
| **Severity** | P0 / P1 / P2 |
| **Reviewers** | @sre-lead, @product-lead |

---

## 1. Summary
[1 paragraf: apa yang terjadi, berapa lama, dampaknya apa]

## 2. Impact
| Metric | Value |
|---|---|
| Duration of outage | 47 minutes |
| Users affected | 12,000 customers |
| Failed requests | 8.2% of total |
| Revenue impact | $3,200 (estimated) |
| Error budget consumed | 65% of monthly budget |

## 3. Timeline (UTC+7)
| Time | Event |
|---|---|
| 03:15 | [Trigger] New deployment checkout-service:v2.4.1 rolled out |
| 03:17 | [Detection] Alert `KubePodCrashLooping` fired |
| 03:18 | [Response] On-call paged |
| 03:22 | [Diagnosis] Exit code 1 — application panic |
| 03:25 | [Mitigation attempt 1] Rollback to v2.4.0 |
| 03:28 | [Problem] Rollback stuck due to PDB not configured |
| 03:35 | [Mitigation attempt 2] Force scale down + recreate |
| 03:45 | [Mitigation attempt 3] Restore from DB snapshot |
| 04:02 | [Recovery] Service restored to 100% |
| 04:30 | [Verification] All SLOs green |

## 4. Root Cause

### Trigger
Deployment of v2.4.1 introduced new code path that required environment variable `PAYMENT_GATEWAY_TIMEOUT` which was missing from ConfigMap.

### Primary Cause
Missing configuration validation in CI pipeline: code that requires env vars can be merged without those vars being defined in any environment.

### Contributing Factors
1. **No pre-deployment smoke test** that would have caught the missing env var
2. **No canary deployment** — pushed directly to 100% traffic
3. **PDB misconfiguration** caused rollback to hang
4. **On-call was not familiar** with the new `force-remove-pod` procedure
5. **Runbook RB-001 was outdated** — referenced kubectl 1.27 syntax, actual cluster runs 1.29

### 5-Whys Analysis
1. **Why** did the service crash? → Missing env var `PAYMENT_GATEWAY_TIMEOUT`
2. **Why** was the env var missing? → ConfigMap was not updated as part of the PR
3. **Why** was ConfigMap not updated? → PR template didn't require config changes
4. **Why** didn't CI catch it? → No integration test that validates env presence
5. **Why** no integration test? → No team priority on defensive coding practices

## 5. Resolution
Restored v2.4.0 from Helm chart history. Manually added `PAYMENT_GATEWAY_TIMEOUT=5000` to ConfigMap for forward compatibility.

## 6. Detection
- ✅ Alert `KubePodCrashLooping` fired within 2 minutes
- ❌ No customer impact alert — first signal was internal

## 7. Response
- On-call responded in 3 minutes (good)
- First mitigation (rollback) took 10 minutes (acceptable)
- Second mitigation (PDB workaround) took 10 minutes (slow)
- Third mitigation (manual recreation) took 17 minutes (too slow)

## 8. What Went Well
- Alert fired within 2 minutes
- On-call responded quickly
- Database integrity preserved (no data loss)

## 9. What Went Wrong
- PDB not configured properly
- Runbook was outdated
- No pre-deployment validation
- Multiple failed rollback attempts

## 10. Where We Got Lucky
- Insiden terjadi jam 3 pagi (low traffic window)
- Customer tidak complain karena checkout-service tidak critical
- Database connection pool masih healthy selama outage

## 11. Action Items
| ID | Action | Owner | Priority | Due Date | Status |
|---|---|---|---|---|---|
| AI-1 | Add env var validation in CI pipeline | @dev-lead | P1 | 2026-08-20 | Open |
| AI-2 | Implement canary deployment (Argo Rollouts) | @sre-lead | P0 | 2026-08-30 | Open |
| AI-3 | Update RB-001 runbook for kubectl 1.29 | @oncall | P2 | 2026-08-15 | Open |
| AI-4 | Configure PDB for all Production deployments | @sre | P1 | 2026-08-25 | Open |
| AI-5 | Add customer-facing error rate alert | @monitoring | P1 | 2026-08-22 | Open |
| AI-6 | Run chaos game day for "missing env var" scenario | @sre | P2 | 2026-09-15 | Open |
```

---

## 3. 🕐 Cara Membuat Timeline dari Observabilitas

Timeline yang akurat adalah **jantung** postmortem. Anda tidak boleh mengandalkan ingatan on-call saja — rekonstruksi dari data:

### 3.1 Detection Time dari Alertmanager

```bash
# Query Alertmanager untuk alert yang firing
$ amtool alert query --alertname=KubePodCrashLooping
Alertname:        KubePodCrashLooping
ActiveAt:         2026-08-10 03:17:24 UTC+7
State:            firing
Labels:           pod=checkout-service-7d8c9b5f8-xxx
Annotations:      summary=Pod crash looping
```

### 3.2 Trigger Time dari Git

```bash
# Cari commit yang men-trigger deployment
$ git log --all --oneline --since="2 hours ago" | head
a3f2e91 deploy: bump checkout-service to v2.4.1  ← trigger!

$ git show a3f2e91 --stat
commit a3f2e91
Author: developer
Date:   2026-08-10 03:15:12 UTC+7

    feat(checkout): add retry logic for payment gateway
```

### 3.3 Error Time dari Loki

```logql
{app="checkout-service"} |= "panic" | json | line_format "{{.timestamp}}: {{.error}}"
2026-08-10 03:15:42: env var PAYMENT_GATEWAY_TIMEOUT required
2026-08-10 03:15:43: panic: runtime error: invalid memory address
```

### 3.4 Recovery Time dari Mimir

```promql
# Plotting error rate over time
sum(rate(http_requests_total{job="checkout-service",code=~"5xx"}[1m]))
# Observe: spike at 03:15, recovers at 04:02
```

### 3.5 Trace untuk Root Cause dari Tempo

Cari span dengan status `error`:
```text
Trace ID: a8f92b1c34891e01
Span: [checkout-service] HTTP POST /api/checkout
  Error: "env var PAYMENT_GATEWAY_TIMEOUT required"
  Duration: 12ms
  Status: ERROR
```

---

## 4. 🔬 Teknik 5-Whys

Teknik dari Toyota Production System untuk menggali **akar masalah sistemik** (bukan akar masalah di permukaan).

### Contoh 5-Whys yang BENAR

**Insiden: Database corruption pada 2026-07-20**

```
Q1: Why database corrupted?
A1: Disk write failed due to bad sector

Q2: Why disk had bad sector?
A2: Disk health not monitored

Q3: Why disk health not monitored?
A3: No SMART monitoring alert setup

Q4: Why no SMART monitoring alert?
A4: SRE team tidak aware tentang kebutuhan monitoring disk health

Q5: Why SRE team tidak aware?
A5: Onboarding checklist tidak menyertakan setup SMART monitoring
   ← ROOT CAUSE SISTEMIK
```

**Action items dari 5-Whys ini:**
- AI-1: Tambah SMART monitoring ke semua server (low-hanging fruit)
- AI-2: Update onboarding checklist untuk menyertakan SMART monitoring
- AI-3: Quarterly disk health audit

### Anti-Pattern 5-Whys yang SALAH

```
Q1: Why service down?
A1: Engineer tidak teliti

Q2: Why tidak teliti?
A2: Lagi ngantuk

Q3: Why ngantuk?
A3: Lembur

Q4: Why lembur?
A4: Project deadline

Q5: Why deadline ketat?
A5: Atasan kejam

→ KESIMPULAN SALAH: "Majikan yang salah" (bukan systemic fix)
→ SEHARUSNYA: cari kenapa deadline ketat (process planning, capacity, dsb)
```

---

## 5. 🎯 Action Items yang SMART

Setiap action item harus memenuhi kriteria **SMART**:

```text
┌──────────────────────────────────────────────────────────────────┐
│ S - Specific    | "Add env validation" bukan "improve CI"        │
│ M - Measurable  | "Block 100% of PRs without env" bukan "kebih │
│                   teliti"                                         │
│ A - Achievable  | Bisa selesai 1 sprint, bukan "rewrite CI"      │
│ R - Relevant    | Berkaitan dengan root cause                     │
│ T - Time-bound  | Due date jelas: 2026-08-20                      │
└──────────────────────────────────────────────────────────────────┘
```

### Contoh Action Items yang Baik vs Buruk

| ❌ Buruk | ✅ Baik |
|---|---|
| "Improve code quality" | "Add `os.Getenv()` validation in 3 critical env vars (P0)" |
| "Be more careful" | "Add pre-commit hook to detect hardcoded secrets (P1)" |
| "Update documentation" | "Update RB-001 with kubectl 1.29 syntax (P2, by 2026-08-15)" |
| "Train team" | "Conduct chaos game day for missing-env scenario (P2, by 2026-09-15)" |

### Tracking Action Items

Buat `postmortem-action-items.md` di repo atau gunakan issue tracker:

```yaml
# .github/postmortems/2026-08-10-pm-001.yaml
postmortem_id: PM-2026-08-001
title: "Checkout service CrashLoopBackOff"
date: 2026-08-10
severity: P1
status: in_progress
action_items:
  - id: AI-1
    title: "Add env var validation in CI"
    owner: "@dev-lead"
    priority: P1
    due: 2026-08-20
    status: open
  - id: AI-2
    title: "Implement canary deployment"
    owner: "@sre-lead"
    priority: P0
    due: 2026-08-30
    status: in_progress
```

---

## 6. 🤝 Postmortem Meeting (Ritual Tim)

### 6.1 Jadwal Meeting

| Severity | Meeting Type | When | Duration | Attendees |
|---|---|---|---|---|
| P0 | Hot wash-up | Segera setelah recovery | 30 min | Semua yang terlibat |
| P1 | Same-day retro | Hari yang sama | 45 min | SRE + Dev + Product |
| P2 | Weekly review | Jumat | 30 min | SRE only |

### 6.2 Agenda Postmortem Meeting

```text
1. Pembukaan (2 min)
   - "Ini adalah blameless meeting. Fokus pada sistem."

2. Timeline walkthrough (10 min)
   - On-call cerita kronologis dengan input dari observability

3. Root cause discussion (10 min)
   - 5-Whys analysis bersama
   - Contributing factors (semua orang boleh menambah)

4. Action items (10 min)
   - Diskusi & assign owner
   - Tentukan due date

5. Lessons learned (5 min)
   - "Apa 1 hal yang akan kita lakukan beda lain kali?"

6. Penutupan (2 min)
   - Catat semua action items
   - Set reminder review
```

### 6.3 Cara Memimpin Meeting yang Blameless

```text
✅ DO:
- "Bisa ceritakan apa yang Anda pikirkan saat itu?"
- "Apa yang sistem Anda tunjukkan saat itu?"
- "Apa yang akan membuat keputusan Anda lebih mudah?"
- "Apakah ada yang melihat warning sign yang kita lewatkan?"

❌ DON'T:
- "Kok bisa salah seperti itu?"
- "Seharusnya kamu cek dulu"
- "Ini kan jelas-jelas kesalahan X"
- "Kapan belajar dari kesalahan?"
```

---

## 7. 🧪 Hands-On: Tulis Postmortem Pertama Anda

### Tugas

Ambil insiden **N+1 Query** atau **CPU Spike** dari Week 10. Buat postmortem lengkap dengan:

1. **Summary** (1 paragraf)
2. **Impact** (kuantitatif: durasi, request affected, error budget)
3. **Timeline** (minimal 8 events dengan timestamp)
4. **Root cause** (trigger + 5-Whys + contributing factors)
5. **5-Whys analysis**
6. **Action items** (minimal 3, SMART, dengan owner)

Upload ke `minggu-11/postmortems/PM-2026-08-XX.md`.

### Template Lokasi File

```text
minggu-11/
├── postmortems/
│   ├── PM-2026-08-001-n1-query.md
│   ├── PM-2026-08-002-cpu-spike.md
│   └── README.md  (index)
└── templates/
    └── postmortem-template.md
```

---

## 8. 📋 Cheat Sheet: Blameless Postmortem Phrases

```text
❌ JANGAN KATA                | ✅ KATA YANG LEBIH BAIK
------------------------------|----------------------------------
"Si X yang salah"             | "Sistem memungkinkan X terjadi"
"Kenapa gak dicek?"          | "Apa yang bisa kita tampilkan agar X tidak perlu menebak?"
"Harusnya lebih teliti"       | "Apa tooling yang bisa otomatis cek ini?"
"User bodoh"                 | "UX terlalu kompleks bagi user"
"Ini jelas human error"      | "5 faktor berkontribusi"
"SpA benar, dev salah"       | "Cross-team collaboration perlu di-improve"
```

---

## 9. ✏️ Latihan Mandiri

1. **Buat postmortem** untuk insiden dari Week 9 atau Week 10 dengan format lengkap.
2. **Tulis 5-Whys** untuk `Network Timeout` insiden dari Week 10.
3. **Definisikan 3 action items** yang SMART dari postmortem di atas.
4. **Buat issue di GitHub/GitLab** untuk setiap action item dengan label `postmortem-action`.
5. **Schedule postmortem meeting** dengan tim Anda untuk review postmortem.

---

## 10. 🔗 Resources Lanjutan

- **Google SRE Book — Chapter 12: Postmortem Culture** (https://sre.google/sre-book/postmortem-culture/)
- **Etsy Debriefing Facilitation Guide** (https://extfiles.etsy.com/DebriefingFacilitationGuide.pdf)
- **Atlassian Incident Handbook** (https://www.atlassian.com/incident-management/handbook)
- **PagerDuty Postmortem Template** (https://www.pagerduty.com/resources/learn/postmortem-template/)

---

**Lanjut ke README Minggu 11:** [README.md](./README.md) — rangkuman semua modul + SRE cheat sheet final.
