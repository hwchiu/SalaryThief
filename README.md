# SalaryThief

BMC collector for physical machines.

設定檔列出每台 server 的 BMC，collector 用 **Redfish** 定期抓 Systems / Chassis / Managers / Thermal / Power，把時間序列打進 **Prometheus**，把庫存快照與異常事件打進 **OpenSearch**，最後用 **Grafana** 看。

本地用 **Kind + DMTF Redfish Mockup Server** 模擬 BMC，不用真的機器就能驗證整條管線。

## Layout

```
cmd/collector/          collector process
internal/redfish/       Redfish HTTP client
internal/collect/       walk service root → snapshot
internal/metrics/       Prometheus gauges (bmc_*)
internal/opensearch/    inventory + events documents
configs/                target inventory (prod + kind)
deploy/kind/            mock BMC + observability stack
hack/kind-up.sh         one-command local lab
```

## Config

`configs/collector.yaml`：

```yaml
targets:
  - name: rack-a-compute-01
    endpoint: https://bmc-compute-01.example.com
    username: ${BMC_USERNAME}
    password: ${BMC_PASSWORD}
    insecure_skip_verify: true
    vendor: dell          # dell | hpe | lenovo | supermicro | dmtf | ...
    labels:
      rack: a
      role: compute
    collect:
      systems: true
      chassis: true
      managers: true
      thermal: true
      power: true
```

密碼用環境變數展開，不要把明文 commit 進 git。

## Metrics

Health encoding: `0=Unknown`, `1=OK`, `2=Warning`, `3=Critical`.

| Metric | Meaning |
|---|---|
| `bmc_up` | last scrape succeeded |
| `bmc_scrape_duration_seconds` | scrape latency |
| `bmc_system_info` | identity labels (serial, model, BIOS) |
| `bmc_system_power_on` | PowerState == On |
| `bmc_system_health` | system health |
| `bmc_system_processor_count` | CPU count |
| `bmc_system_memory_gib` | memory |
| `bmc_chassis_health` | chassis health |
| `bmc_manager_info` / `bmc_manager_health` | BMC firmware + health |
| `bmc_temperature_celsius` | sensors |
| `bmc_fan_reading` | fans |
| `bmc_power_consumed_watts` | chassis power |
| `bmc_psu_health` | PSU |

OpenSearch indices:

- `bmc-inventory` — one document per target per scrape (full snapshot)
- `bmc-events` — only non-OK health and scrape failures

## Local Kind lab

Needs `docker`, `kind`, `kubectl`.

```bash
make test
make kind-up
```

Then:

| Service | URL |
|---|---|
| Grafana | http://127.0.0.1:30000 (`admin` / `admin`) |
| Prometheus | http://127.0.0.1:30090 |
| OpenSearch | http://127.0.0.1:30920 |

Dashboard: **SalaryThief BMC**. OpenSearch datasource is provisioned as `bmc-*`.

Tear down:

```bash
make kind-down
```

## Run collector only

```bash
go run ./cmd/collector -config configs/collector.yaml
curl localhost:9100/metrics
```

Against a single mock BMC without Kind:

```bash
docker run --rm -p 8000:8000 docker.io/dmtf/redfish-mockup-server:latest
# point a target endpoint at http://127.0.0.1:8000
```

## Next

- Session-based Redfish auth (some BMCs cap Basic-auth connections)
- EventService subscription instead of scrape-only events
- Storage / NetworkAdapters / LogServices collectors
- Per-vendor quirks (`vendor: dell` thermal paths, iLO power)
- Recording rules + alertmanager for `bmc_up == 0` and `health >= 2`
