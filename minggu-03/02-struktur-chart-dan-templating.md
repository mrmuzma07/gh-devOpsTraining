# Modul 02: Anatomi Helm Chart & Sintaks Templating

> **Target Pembelajaran:** Memahami struktur direktori Helm Chart (`Chart.yaml`, `values.yaml`, `templates/`, `_helpers.tpl`) serta menguasai sintaks Go Templating pada Helm.

---

## 1. Anatomi Struktur Folder Helm Chart

Sebuah Helm Chart yang lengkap memiliki struktur folder sebagai berikut:

```text
my-chart/
├── Chart.yaml          # Metadata utama Chart (Nama, Versi, Deskripsi)
├── values.yaml         # Nilai konfigurasi default variabel
├── values-dev.yaml     # (Opsional) Override nilai variabel untuk Env Dev
├── values-prod.yaml    # (Opsional) Override nilai variabel untuk Env Prod
├── .helmignore         # Daftar file/folder yang diabaikan saat memaketkan Chart
└── templates/          # Berkas template Kubernetes YAML
    ├── _helpers.tpl    # Snippet/fungsi pembantu yang reusable (Named Templates)
    ├── deployment.yaml
    ├── service.yaml
    ├── configmap.yaml
    ├── secret.yaml
    └── pvc.yaml
```

---

## 2. Berkas Metadata: `Chart.yaml`

File `Chart.yaml` berisi informasi identitas paket Chart:

```yaml
apiVersion: v2
name: go-app
description: Helm Chart untuk Aplikasi Go Production-Ready
type: application
version: 0.1.0        # Versi paket Helm Chart itu sendiri
appVersion: "1.0.0"   # Versi aplikasi (Go App) yang dikemas
```

> **Beda `version` vs `appVersion`:**
> - `version`: Nomor versi paket Chart. Naik jika Anda mengubah template YAML di Chart.
> - `appVersion`: Nomor versi aplikasi Go. Naik jika Anda meng-update versi Docker image aplikasi.

---

## 3. Sintaks Templating Helm (Go Text/Template)

Helm menggunakan engine **Go text/template** yang diidentifikasikan dengan tanda kurung kurawal ganda `{{ ... }}`.

### Object Bawaan Helm Utama:
1. **`.Values`**: Mengakses variabel yang didefinisikan di `values.yaml`.
   - Contoh: `{{ .Values.replicaCount }}`
2. **`.Release`**: Mengakses informasi release yang sedang berjalan.
   - Contoh: `{{ .Release.Name }}` *(nama release, misal: go-app-dev)*
   - Contoh: `{{ .Release.Namespace }}`
3. **`.Chart`**: Mengakses isi dari `Chart.yaml`.
   - Contoh: `{{ .Chart.Name }}`

---

## 4. Fungsi & Pipe Operator (`|`)

Helm menyediakan puluhan fungsi manipulasi teks yang dipanggil menggunakan *pipe operator* (`|`).

### Contoh Fungsi Populer:

1. **`quote` (Menambahkan tanda petik dua):**
   ```yaml
   env: {{ .Values.envName | quote }}
   # Hasil: env: "production"
   ```

2. **`default` (Nilai cadangan jika variabel kosong):**
   ```yaml
   imagePullPolicy: {{ .Values.image.pullPolicy | default "IfNotPresent" }}
   ```

3. **`nindent` (New line + Auto Indentasi YAML):**
   ```yaml
   resources:
     {{- toYaml .Values.resources | nindent 4 }}
   ```
   *(Tanda dash `-` di `{{-` berfungsi menghapus spasi/newline kosong sebelum blok)*.

---

## 5. Named Templates (`_helpers.tpl`)

File `_helpers.tpl` digunakan untuk membuat potongan logika/label yang sering digunakan kembali (*reusable snippet*) di berbagai file template.

### Contoh Definisi di `_helpers.tpl`:
```gotemplate
{{/* Generate nama lengkap resource */}}
{{- define "go-app.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
```

### Cara Memanggilnya di `deployment.yaml`:
```yaml
metadata:
  name: {{ include "go-app.fullname" . }}
```

---

## Ringkasan Modul 02

- **`Chart.yaml`** mencatat metadata versi; **`values.yaml`** menyimpan variabel default.
- Gunakan `{{ .Values.path.key }}` untuk membaca nilai konfigurasi.
- **Pipe (`\|`)** dan **`nindent`** membantu memformat inden YAML secara presisi tanpa error sintaks.
- **`_helpers.tpl`** memusatkan pembuatan nama dan label resource agar konsisten di seluruh cluster.
