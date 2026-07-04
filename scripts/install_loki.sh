#!/bin/bash
set -e

echo "Downloading Loki and Promtail..."
cd ~
mkdir -p grafana-logs
cd grafana-logs

wget -qO loki.zip "https://github.com/grafana/loki/releases/download/v2.9.4/loki-linux-arm64.zip"
wget -qO promtail.zip "https://github.com/grafana/loki/releases/download/v2.9.4/promtail-linux-arm64.zip"

unzip -o loki.zip
unzip -o promtail.zip

chmod a+x loki-linux-arm64 promtail-linux-arm64

echo "Creating Loki config..."
cat << 'CFG' > loki-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  instance_addr: 127.0.0.1
  path_prefix: /data/data/com.termux/files/home/grafana-logs/loki
  storage:
    filesystem:
      chunks_directory: /data/data/com.termux/files/home/grafana-logs/loki/chunks
      rules_directory: /data/data/com.termux/files/home/grafana-logs/loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://localhost:9093
CFG

echo "Creating Promtail config..."
cat << 'CFG2' > promtail-config.yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /data/data/com.termux/files/home/grafana-logs/positions.yaml

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
- job_name: pm2_logs
  static_configs:
  - targets:
      - localhost
    labels:
      job: pm2
      __path__: /data/data/com.termux/files/home/.pm2/logs/*.log
CFG2

echo "Starting Loki and Promtail with PM2..."
pm2 start ./loki-linux-arm64 --name="loki" -- -config.file=loki-config.yaml
pm2 start ./promtail-linux-arm64 --name="promtail" -- -config.file=promtail-config.yaml
pm2 save

echo "Done!"
