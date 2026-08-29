# SalaryThief v2 — Design Handoff

## Purpose

This document transfers the architectural decisions already made for SalaryThief so a local coding agent can continue implementation without re-litigating the core design.

SalaryThief is a Redfish/BMC observability collector for physical servers. The existing repository is considered a useful MVP/prototype and should be evolved rather than replaced.

## Production context

The target environment has multiple Kubernetes environments. Each environment has its own Prometheus and roughly 200 physical servers. Only a central collector host/service is allowed to reach the BMC network.

The observability stack available is:

- Prometheus
- OpenSearch or Elasticsearch
- Grafana

There is no requirement for a single global Grafana view across every environment. Each environment/dashboard is scoped to the relevant Prometheus datasource and its servers.

## Core architecture decision

Prometheus MUST NOT trigger Redfish polling synchronously.

Rejected design:

```text
Prometheus scrape
      ↓
SalaryThief
      ↓
BMC request
      ↓
return /metrics
```

Required design:

```text
Background Scheduler
        ↓
   bounded workers
        ↓
       BMCs
        ↓
 normalized cache/state
        ↓
     /metrics
        ↓
   Prometheus
```

### Why

BMCs are slow management-plane devices. Their latency, firmware defects, transient network failures, or authentication failures must not sit in the synchronous Prometheus scrape path.

This design also prevents multiple Prometheus instances from independently hitting the same BMC fleet and fits the network requirement that only the central collector can reach BMC endpoints.

## Data ownership

### Prometheus

Prometheus is for cheap, regularly sampled, graphable, alertable signals:

- BMC availability
- collection latency/errors
- overall hardware health
- temperatures
- fan speed
- power
- CPU telemetry when available
- memory ECC counters when available
- drive health/temperature/media errors/predictive failure when available
- collector self-observability

### OpenSearch / Elasticsearch

OpenSearch/Elasticsearch is for document-oriented and higher-cardinality data:

- hardware inventory
- Redfish LogService/EventLog events
- inventory changes
- raw sanitized events when useful
- future correlated incidents

Inventory should be written on refresh/change, not every telemetry cycle.

## Scope identity

Use an explicit scope field such as:

```text
observability_scope=k8s-prod-a
```

A server registry entry can conceptually contain:

```yaml
server_id: server-042
bmc_endpoint: https://10.x.x.x
observability_scope: k8s-prod-a
prometheus_scope: prometheus-k8s-prod-a
credential_ref: bmc-credential-dell
```

Credentials must remain external to metrics/logs/OpenSearch.

## Component identity versus location

Physical location and hardware identity must remain separate.

Example:

```text
component_id = drive-8f92ab
location     = Bay 03
serial       = AAA
```

After replacement:

```text
location     = Bay 03
serial       = BBB
```

The system must be able to express that Bay 03 stayed the same physical location while the hardware identity changed.

Expected inventory change document:

```json
{
  "event_type": "inventory_change",
  "server": "server-042",
  "component": "drive",
  "location": "Bay 03",
  "change": "replaced",
  "old_serial": "AAA",
  "new_serial": "BBB"
}
```

Supported change types should eventually include:

- added
- removed
- replaced
- firmware_changed
- capacity_changed
- model_changed

## Desired physical topology model

```text
Server
├── CPU Socket 0
├── CPU Socket 1
├── Memory
│   ├── DIMM A1
│   ├── DIMM A2
│   ├── DIMM B1
│   └── DIMM B2
├── Storage Controller
│   └── Enclosure 1
│       ├── Bay 01
│       ├── Bay 02
│       ├── Bay 03
│       └── Bay 04
├── PSU
│   ├── PSU1
│   └── PSU2
├── Fan
├── NIC
└── GPU
```

The system should retain enough normalized inventory to support RMA evidence: slot/bay, serial, part number, manufacturer/model, capacity, firmware, health, and timestamps.

## Event model

MVP event source: Redfish LogService/EventLog incremental polling.

EventService subscription is a future enhancement and is not required to make the MVP production architecture valid.

Normalized event example:

```json
{
  "@timestamp": "2026-08-29T10:30:12Z",
  "observability_scope": "k8s-prod-a",
  "server": "server-042",
  "vendor": "Dell",
  "model": "PowerEdge R760",
  "severity": "critical",
  "category": "storage",
  "event_type": "media_failure",
  "component": "drive",
  "component_id": "drive-03",
  "location": "Bay 03",
  "event_id": "12345",
  "event_fingerprint": "abcdef",
  "source": "Redfish",
  "redfish_resource": "/redfish/v1/Systems/...",
  "message": "Drive failure"
}
```

Dedup preference:

```text
server + source + event_id
```

Fallback fingerprint:

```text
server + component + location + event_type + message
```

Normalized severities:

```text
OK INFO WARNING CRITICAL UNKNOWN
```

Normalized categories:

```text
storage memory processor power thermal fan network gpu pcie firmware bmc system security unknown
```

## Failure semantics

Differentiate:

- BMC unreachable
- timeout
- auth failure
- HTTP error
- TLS failure
- schema error
- partial resource failure

`redfish_up=0` means the collector cannot currently retrieve usable BMC data. It does NOT mean the physical server hardware is failed.

If the BMC is unreachable:

```text
connectivity = DOWN
hardware     = UNKNOWN
```

If one subtree fails, unrelated data remains valid. Example:

```text
BMC        UP
CPU        OK
Memory     OK
Storage    UNKNOWN/ERROR
```

Freshness must be visible through timestamps/data age.

## Scheduling guidance

Initial conceptual intervals:

```text
BMC availability         15–30s
overall health           30s
temperature              30s
fan                      30–60s
power                    30–60s
CPU telemetry            30–60s
memory ECC               30–60s
drive health/temp        30–60s
EventLog                  15–60s incremental
inventory                6h
firmware                  6–24h
```

Inventory refresh may also be triggered on startup, server/BMC reboot, hardware-change events, detected mismatch, or manual refresh.

## Concurrency and backpressure

Use bounded queues and workers:

```text
scheduler → work queue → workers → BMC
```

Initial tunable worker counts may look like:

```yaml
workers:
  telemetry: 32
  inventory: 8
  events: 16
```

These are starting points only and must be benchmarked.

Priority order:

1. availability / overall health / critical events
2. temperature / power / ECC / drive health
3. detailed telemetry
4. inventory / firmware

Repeatedly failing BMCs should back off with jitter and may temporarily stop expensive detail polling while retaining cheap health probes.

Telemetry, event writes, and inventory writes must not block each other. An OpenSearch outage must not break `/metrics`.

## Prometheus label guidance

Recommended common dimensions:

```text
observability_scope
server
component
component_id
location
```

Avoid using the following as ordinary labels:

```text
serial
part_number
manufacturer
model
firmware
service_tag
asset_tag
```

Identity belongs primarily in inventory documents.

## Collector self-observability

Useful metrics include:

```text
redfish_collector_bmcs_total
redfish_collector_bmc_up
redfish_collector_requests_total
redfish_collector_request_errors_total
redfish_collector_request_duration_seconds
redfish_collector_events_total
redfish_collector_inventory_total
redfish_collector_queue_depth
redfish_collector_workers_active
redfish_collector_last_success_timestamp
redfish_collector_bmc_timeouts_total
redfish_collector_bmc_auth_failures_total
redfish_collector_schema_errors_total
redfish_collector_event_deduplicated_total
redfish_collector_inventory_changes_total
```

BMC-level freshness/collection metrics should include equivalents of:

```text
redfish_up
redfish_collection_duration_seconds
redfish_collection_errors_total
redfish_last_success_timestamp_seconds
redfish_data_age_seconds
```

## Long-term incident correlation

A future incident layer should correlate related events such as:

```text
Drive Temperature High
→ Media Error
→ Predictive Failure
→ Drive Failure
→ RAID Degraded
```

into one incident. This is Phase 2+ and must not block the MVP.

## Non-goals

SalaryThief is not an OS/application telemetry collector. Keep these out of scope:

- disk IOPS
- throughput
- latency
- queue depth
- raw SMART attributes
- filesystem/mount/inode metrics
- application SLIs
- write/control operations against server power/boot/virtual media
