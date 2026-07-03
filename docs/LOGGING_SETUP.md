# Panduan Instalasi Loki + Grafana (Server Live / Native)

Panduan ini untuk menginstal Loki, Promtail, dan Grafana secara **native** (tanpa Docker) di server Linux produksi.

---

## Prasyarat

- Server Linux (Ubuntu/Debian recommended)
- Akses root atau sudo
- Backend Go Siakad sudah dikonfigurasi untuk menulis log ke file

## 1. Konfigurasi Backend Go

Tambahkan variabel `LOG_FILE_PATH` di file `.env` backend:

```env
LOG_FILE_PATH=/var/log/siakad/app.log
```

Buat direktori log:

```bash
sudo mkdir -p /var/log/siakad
sudo chown $USER:$USER /var/log/siakad
```

---

## 2. Instalasi Loki

### Download Binary

```bash
# Cek versi terbaru di https://github.com/grafana/loki/releases
LOKI_VERSION="3.5.0"  # Ganti dengan versi terbaru

cd /tmp
curl -LO "https://github.com/grafana/loki/releases/download/v${LOKI_VERSION}/loki-linux-amd64.zip"
unzip loki-linux-amd64.zip
sudo mv loki-linux-amd64 /usr/local/bin/loki
sudo chmod +x /usr/local/bin/loki
```

### Buat File Konfigurasi

```bash
sudo mkdir -p /etc/loki
sudo mkdir -p /var/lib/loki
```

Buat file `/etc/loki/config.yml`:

```yaml
auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /var/lib/loki
  storage:
    filesystem:
      chunks_directory: /var/lib/loki/chunks
      rules_directory: /var/lib/loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2020-10-24
      store: tsdb
      object_store: filesystem
      schema: v13
      index:
        prefix: index_
        period: 24h

limits_config:
  allow_structured_metadata: true
  volume_enabled: true
```

### Buat Systemd Service

Buat file `/etc/systemd/system/loki.service`:

```ini
[Unit]
Description=Loki Log Aggregation System
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/loki -config.file=/etc/loki/config.yml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable loki
sudo systemctl start loki

# Verifikasi
sudo systemctl status loki
curl -s http://localhost:3100/ready
```

---

## 3. Instalasi Promtail

### Download Binary

```bash
LOKI_VERSION="3.5.0"  # Gunakan versi yang sama dengan Loki

cd /tmp
curl -LO "https://github.com/grafana/loki/releases/download/v${LOKI_VERSION}/promtail-linux-amd64.zip"
unzip promtail-linux-amd64.zip
sudo mv promtail-linux-amd64 /usr/local/bin/promtail
sudo chmod +x /usr/local/bin/promtail
```

### Buat File Konfigurasi

Buat file `/etc/promtail/config.yml`:

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  - job_name: siakad-backend
    static_configs:
      - targets:
          - localhost
        labels:
          app: siakad
          environment: production
          __path__: /var/log/siakad/*.log
```

### Buat Systemd Service

Buat file `/etc/systemd/system/promtail.service`:

```ini
[Unit]
Description=Promtail Log Agent
After=network.target loki.service

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/promtail -config.file=/etc/promtail/config.yml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo mkdir -p /etc/promtail
sudo systemctl daemon-reload
sudo systemctl enable promtail
sudo systemctl start promtail

# Verifikasi
sudo systemctl status promtail
```

---

## 4. Instalasi Grafana

### Tambah Repository & Install

```bash
# Tambah GPG key
sudo apt-get install -y apt-transport-https software-properties-common wget
sudo mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null

# Tambah repository
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list

# Install
sudo apt-get update
sudo apt-get install grafana -y
```

### Konfigurasi Datasource Loki

Buat file `/etc/grafana/provisioning/datasources/loki.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://localhost:3100
    isDefault: true
    jsonData:
      maxLines: 1000
```

### Jalankan Grafana

```bash
sudo systemctl daemon-reload
sudo systemctl enable grafana-server
sudo systemctl start grafana-server

# Verifikasi
sudo systemctl status grafana-server
```

### Akses Grafana

Buka browser: `http://IP-SERVER-ANDA:3000`

- **Username:** admin
- **Password:** admin (akan diminta ganti saat login pertama)

---

## 5. Keamanan (PENTING untuk Server Produksi!)

### Ganti Password Grafana

Segera ganti password admin default setelah login pertama kali.

### Batasi Akses dengan Firewall

Grafana, Loki, dan Promtail sebaiknya **TIDAK** terbuka ke internet publik. Batasi akses hanya dari IP Anda:

```bash
# Hanya izinkan Grafana dari IP tertentu
sudo ufw allow from IP_ANDA_DI_RUMAH to any port 3000

# Blokir akses publik ke Loki & Promtail
sudo ufw deny 3100
sudo ufw deny 9080
```

Atau jika Anda menggunakan Nginx sebagai reverse proxy, tambahkan autentikasi:

```nginx
location /grafana/ {
    proxy_pass http://localhost:3000/;
    auth_basic "Restricted";
    auth_basic_user_file /etc/nginx/.htpasswd;
}
```

---

## 6. Contoh Query LogQL di Grafana

Setelah Grafana berjalan, buka menu **Explore** > pilih datasource **Loki**, lalu coba query berikut:

### Melihat Semua Log Siakad
```
{app="siakad"}
```

### Filter Hanya Error
```
{app="siakad"} | json | level="ERROR"
```

### Filter Per Modul (misal: response error)
```
{app="siakad"} | json | msg="Response error"
```

### Mencari Error Tertentu
```
{app="siakad"} | json | message=~".*database.*"
```

### Melihat HTTP Request yang Lambat (> 500ms)
```
{app="siakad"} | json | msg="http_request" | duration_ms > 500
```

### Melihat Request dari User Tertentu
```
{app="siakad"} | json | user_id="12"
```

---

## 7. Troubleshooting

### Log Tidak Muncul di Grafana?

1. Pastikan file log ada: `ls -la /var/log/siakad/app.log`
2. Pastikan Promtail berjalan: `sudo systemctl status promtail`
3. Cek log Promtail: `sudo journalctl -u promtail -f`
4. Pastikan Loki berjalan: `curl http://localhost:3100/ready`

### Grafana Tidak Bisa Diakses?

1. Cek status: `sudo systemctl status grafana-server`
2. Cek port: `sudo ss -tlnp | grep 3000`
3. Cek firewall: `sudo ufw status`
