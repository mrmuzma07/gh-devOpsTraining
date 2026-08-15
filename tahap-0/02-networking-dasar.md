# Modul 02 — Networking Dasar

## Tujuan

Memahami perjalanan sebuah request sehingga istilah seperti DNS, port, Service, Ingress, timeout, dan connection refused tidak terasa seperti kotak hitam.

## Materi Inti

### 1. Identitas dan jalur jaringan

- IPv4/IPv6, interface, MAC, subnet, gateway, route, dan loopback (`127.0.0.1`/`::1`).
- CIDR, private network, NAT, firewall, proxy, dan perbedaan koneksi lokal dengan koneksi antar-host.
- Perintah inspeksi: `ip addr`, `ip route`, `ip neigh`, `ping`, dan `traceroute` atau `tracepath`.

### 2. DNS

- Resolver, recursive query, record A/AAAA, CNAME, NS, TTL, dan `/etc/hosts`.
- Bedakan kegagalan resolusi nama dari kegagalan koneksi ke IP.

```bash
dig example.com
getent hosts example.com
cat /etc/resolv.conf
```

### 3. Port dan protokol

- TCP handshake, connection reset, timeout, UDP, listening socket, dan ephemeral port.
- Port hanya alamat service pada sebuah host; port terbuka tidak otomatis berarti aplikasi sehat.
- HTTP method, status code, header, body, keep-alive, TLS, dan sertifikat.

```bash
curl -vI https://example.com
curl -sS -o /dev/null -w 'status=%{http_code} time=%{time_total}s\n' https://example.com
ss -tulpen
```

### 4. Hubungan dengan Kubernetes

- Pod IP bersifat sementara; Service memberi discovery dan alamat virtual yang stabil.
- Ingress atau Gateway menerima HTTP(S) dari luar dan meneruskannya ke Service.
- NetworkPolicy mengatur siapa yang boleh berkomunikasi; DNS dan policy harus diverifikasi saat troubleshooting.

## Latihan

1. Gambar alur `client → DNS → IP:port → TCP/TLS → HTTP service` untuk satu endpoint.
2. Uji endpoint dengan `curl -v`, catat status code, latency, alamat IP, dan sertifikat secara aman.
3. Bandingkan hasil ketika nama host valid, nama host tidak ada, port tidak listening, dan endpoint lambat.
4. Pada cluster lab Tahap 1, gunakan `kubectl get svc`, `kubectl get endpoints`, dan `kubectl describe ingress` untuk menghubungkan teori dengan resource Kubernetes.

## Checklist

- [ ] Dapat menjelaskan perbedaan DNS failure, timeout, refused, reset, dan HTTP 5xx.
- [ ] Dapat membaca route, listening port, dan hasil `curl -v`.
- [ ] Dapat menjelaskan fungsi Service dan Ingress tanpa menyamakan keduanya.
- [ ] Dapat menguji konektivitas tanpa menjalankan port scan terhadap sistem yang tidak dimiliki.
- [ ] Dapat menyebutkan data apa yang perlu dicatat ketika melaporkan masalah jaringan.

## Batasan praktik

Lakukan pengujian hanya pada host, endpoint, dan cluster yang Anda miliki atau yang memang disediakan untuk latihan. Jangan memasukkan token, cookie, atau payload sensitif ke command yang disimpan di history.
