# Minggu 5 — Metrics & Observability (Grafana Stack)

Selamat datang di materi pembelajaran **Minggu 5: Metrics & Observability**.

Di minggu ini, kita membangun pilar observability pertama: **Metrics**. Anda akan mempelajari format metrik Prometheus, menginstal **Grafana Mimir** sebagai storage terpusat, mengonfigurasi **Grafana Alloy** sebagai agen collector, serta membuat dashboard visualisasi interaktif di **Grafana**.

---

## 🗺️ Struktur & Daftar Modul Pembelajaran

Materi Minggu 5 dibagi menjadi 5 modul dokumen dan berkas manifest pendukung:

| No | Dokumen Modul | Deskripsi Topik | Jenis |
| :---: | :--- | :--- | :---: |
| 01 | [Modul 01: Konsep Metrics & Prometheus](./01-konsep-telemetry-dan-metrics-prometheus.md) | 3 Pilar Observability, Format Prometheus, Counter/Gauge/Histogram, Scraping vs Remote Write | 📖 Teori |
| 02 | [Modul 02: Arsitektur Grafana Stack](./02-arsitektur-grafana-alloy-dan-mimir.md) | Arsitektur Grafana, Alloy Collector Agent, & Mimir Storage | 📖 Teori |
| 03 | [Modul 03: Metric Exporters & PromQL](./03-exporter-dan-scraping-target.md) | Node Exporter, kube-state-metrics, cAdvisor, Go App instrumentation, & Sintaks PromQL | 📖 Teori |
| 04 | [Modul 04: Lab Deploy Observability Stack](./04-lab-install-observability-stack-dan-dashboard.md) | Hands-on deploy Mimir + Grafana + Alloy & import Dashboard JSON | 🧪 Lab |
| 05 | [Modul 05: Lab Incident CPU Stress](./05-lab-incident-cpu-stress-monitoring.md) | Simulasi beban CPU stress test & live monitoring di Dashboard Grafana | 🧪 Lab |
| 📂 | [Observability Manifests `manifests/`](./manifests/) | Manifest `01-mimir.yaml`, `02-grafana.yaml`, `03-alloy.yaml` | 💻 Code |
| 📂 | [Grafana Dashboards `dashboards/`](./dashboards/) | Berkas JSON dashboard `cluster-and-app-dashboard.json` | 💻 Code |

---

## 🎯 Target & Checklist Capaian Pembelajaran

- [ ] Memahami 3 pilar utama Observability (**Metrics, Logs, Traces**).
- [ ] Memahami perbedaan tipe metrik **Counter**, **Gauge**, **Histogram**, dan **Summary**.
- [ ] Memahami peran **Grafana Alloy** sebagai collector dan **Grafana Mimir** sebagai storage.
- [ ] Mampu menulis query **PromQL** sederhana (fungsi `rate()`, agregasi `sum`).
- [ ] Berhasil menginstal Mimir, Grafana, dan Alloy di cluster k3s.
- [ ] Berhasil mengimpor Grafana Dashboard JSON dan melihat grafik metrik real-time.
- [ ] Mampu mendeteksi lonjakan CPU (*CPU Stress Spike*) melalui visualisasi Grafana Dashboard.

---

## 🚀 Ringkasan Alur Telemetri Metrics

```
  [Node Exporter / kube-state-metrics / Go App]
                      │
                      ▼ (Scrape HTTP /metrics per 5s)
               [Grafana Alloy Agent]
                      │
                      ▼ (Remote Write Protobuf)
              [Grafana Mimir Storage]
                      │
                      ▼ (PromQL Query)
           [Grafana Dashboard Front-End]
```

---

## ➡️ Perjalanan Selanjutnya (Persiapan Minggu 6)

Setelah menguasai metrik kuantitatif, Anda siap masuk ke **Minggu 6: Logging (Grafana Loki)**, di mana kita akan mempelajari:
- Pilar Observability kedua: **LOGS**.
- Menginstal **Grafana Loki** (Penyimpanan Log Terpusat).
- Mengonfigurasi **Grafana Alloy** untuk mengumpulkan log `stdout`/`stderr` dari seluruh container Pod.
- Menggunakan bahasa query **LogQL** untuk mencari pesan error aplikasi!
