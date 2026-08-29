# Architecture

```
                  ┌─────────────────────────────┐
  physical BMCs   │ iDRAC / iLO / XCC / BMC     │
  or Kind mocks   │ Redfish HTTPS :443 / :8000  │
                  └──────────────┬───────────────┘
                                │ GET /redfish/v1/{Systems,Chassis,Managers}
                                ▼
                  ┌─────────────────────────────┐
                  │ SalaryThief collector       │
                  │  - YAML target inventory    │
                  │  - periodic scrape          │
                  │  - /metrics  :9100          │
                  │  - inventory + events JSON  │
                  └───────┬──────────────┬──────┘
                          │              │
                          ▼              ▼
                 Prometheus         OpenSearch
                 time series        documents
                          │              │
                          └──────┬───────┘
                                 ▼
                              Grafana
```

## Why two backends

| Question | Store |
|---|---|
| Is this BMC reachable right now? | Prometheus `bmc_up` |
| How hot is CPU1 over the last 6h? | Prometheus `bmc_temperature_celsius` |
| What is the serial / SKU / BIOS of this box? | OpenSearch `bmc-inventory` |
| When did PSU2 go Critical? | OpenSearch `bmc-events` |

Prometheus is the wrong place for large, slowly changing inventory JSON.
OpenSearch is the wrong place for cheap 15s gauge scrapes. The collector writes both.

## Collection model

The collector owns the target list. Prometheus scrapes *the collector*, not each BMC.
That keeps BMC session pressure under control and matches how most hardware teams
actually operate: one inventory file, many machines.

Each scrape walks the Redfish service root and follows `@odata.id` links.
Vendor-specific quirks should be isolated later behind `vendor:` on the target
instead of hard-coding iDRAC vs iLO URLs.

## Kind lab

`hack/kind-up.sh` brings up:

- two `dmtf/redfish-mockup-server` pods (public-rackmount1 mock)
- collector pointed at those Services
- Prometheus, OpenSearch (security plugin off), Grafana

This is for pipeline verification, not for firmware-accurate vendor behaviour.
For that, keep a staging iDRAC/iLO in `configs/collector.yaml`.
