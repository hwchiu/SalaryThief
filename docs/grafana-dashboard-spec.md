# Grafana Dashboard Implementation Specification

## Product model

Grafana is an operational troubleshooting UI, not the source of truth.

```text
Prometheus  → live health / telemetry / freshness
OpenSearch  → inventory / events / hardware identity
Grafana     → correlate and navigate
```

Each Grafana deployment/datasource represents one observability environment. Do not add a global `$cluster` variable unless deployment architecture later changes.

## Shared variables

### `$server`
Source: Prometheus label values from a stable server-level metric such as `redfish_up`.

Single or multi-select depending on dashboard.

### `$component`
Values should remain bounded:

```text
system
processor
memory
storage
power
thermal
fan
network
gpu
pcie
bmc
```

### Time range
Use Grafana native time picker.

## State mapping

Hardware state:

```text
UNKNOWN  = 0
OK       = 1
WARNING  = 2
CRITICAL = 3
ERROR    = 4
```

Connectivity is separate:

```text
redfish_up 1 = reachable/current core collection
redfish_up 0 = unavailable
```

Do not visually present `redfish_up=0` as hardware CRITICAL.

Staleness should be visually distinct from hardware failure.

---

# Dashboard 01 — Physical Server Overview

## Purpose

Answer within seconds:

```text
How many servers are healthy?
Which servers need attention?
Is the problem hardware, BMC connectivity, or stale collection?
Where should I click next?
```

## Variables

Optional `$server` multi-select for filtering; default All.

## Row 1 — Fleet KPIs

Panels:

```text
Total Servers
Healthy
Warning
Critical
BMC Offline
Stale Data
```

### Total Servers

Intent:

```promql
count(redfish_up)
```

or equivalent stable one-series-per-server metric.

### BMC Offline

```promql
sum(redfish_up == 0)
```

### Stale Data

Use a configurable dashboard threshold, initially e.g. 120s:

```promql
sum(redfish_data_age_seconds > 120)
```

Do not equate stale with offline.

### Healthy / Warning / Critical

Prefer a normalized server overall-health metric if available.

Do not infer whole-server CRITICAL merely from connectivity.

## Row 2 — Server Health Matrix

This is the primary fleet panel.

Recommended visualization: table/state timeline/status grid.

Rows:

```text
server-001
server-002
...
```

Columns:

```text
BMC
System
CPU
Memory
Storage
Power
Thermal
Fan
Freshness
```

Cell states:

```text
OK
WARNING
CRITICAL
UNKNOWN
ERROR
OFFLINE
STALE
```

Click server → Dashboard 02 with `$server` populated.

This panel is more important than decorative charts.

## Row 3 — Active Problems

Panels:

```text
Servers requiring attention
Problem count by component
Stale/offline targets
```

"Servers requiring attention" should expose:

```text
server
highest severity
component
location if available
data age
```

Prometheus should supply live status. Location may be enriched through OpenSearch only where practical.

## Row 4 — Recent Hardware Events

Datasource: OpenSearch.

Columns:

```text
timestamp
server
severity
category
component
location
message
```

Default filter:

```text
warning OR critical
```

Click event → Dashboard 03, preserving server/time context.

## Acceptance

An operator must be able to identify:

```text
server-042
Storage CRITICAL
Bay 03
```

or:

```text
server-087
BMC OFFLINE
hardware health UNKNOWN
```

without confusing those two cases.

---

# Dashboard 02 — Server Hardware Detail

## Purpose

Answer:

```text
What exactly is wrong with this server?
Where is the component physically?
What telemetry led to this state?
What exact part is installed there?
```

## Variable

`$server` required, single select.

## Header

Show:

```text
server
connectivity
overall health
last success
data age
vendor
model
service tag / asset identity
```

Prometheus supplies connectivity/freshness.
OpenSearch supplies identity.

## Row — CPU

Panels:

```text
socket health cards
CPU temperature trends
CPU power trends
```

Inventory table:

```text
socket
manufacturer
model
serial if available
part number if available
cores
threads
firmware/microcode if available
```

Do not add serial/model as Prometheus labels merely to populate this table. Query OpenSearch.

## Row — Memory

Primary visualization: DIMM slot table/map.

Columns:

```text
slot/location
health
capacity
speed
manufacturer
part number
serial
correctable ECC
uncorrectable ECC
```

Example:

```text
DIMM A1  OK
DIMM A2  CRITICAL
DIMM A3  OK
```

Clicking DIMM A2 should preserve location context for events.

## Row — Storage

Topology must be visible:

```text
Controller
  └─ Enclosure
      ├─ Bay 01
      ├─ Bay 02
      ├─ Bay 03
      └─ Bay 04
```

Minimum table:

```text
controller
enclosure
bay/location
health
predictive failure
temperature
media errors
manufacturer
model
serial
part number
firmware
capacity
```

Important:

```text
Bay 03 = physical location
Serial AAA = current hardware identity
```

Never merge them into one identity.

A replacement should allow the same Bay 03 to later display Serial BBB.

## Row — Power

Panels:

```text
current system power
power trend
PSU health
input/output watts
```

PSU table:

```text
slot
health
manufacturer
model
serial
part number
capacity
firmware
```

## Row — Thermal / Fans

Panels:

```text
inlet temperature
temperature sensors
fan RPM
thermal health
fan health
```

Prefer useful sensor grouping rather than rendering hundreds of anonymous lines.

## Row — Network / GPU / PCIe

May be collapsed by default in MVP.

Tables should preserve:

```text
physical location
health
model
firmware
identity
```

## Row — Recent Events

OpenSearch event table scoped to `$server`.

Click → Dashboard 03.

## Row — Inventory changes

Phase 2.

Show:

```text
timestamp
component
location
change
old identity
new identity
```

---

# Dashboard 03 — Hardware Events & Timeline

## Purpose

Answer:

```text
What happened?
When?
What component/location?
What telemetry changed around the same time?
```

## Variables

```text
$server
$component
$severity
```

## Row 1 — Event summary

Stats:

```text
critical events
warning events
unique affected components
```

## Row 2 — Timeline

Datasource: OpenSearch.

Event fields:

```text
timestamp
severity
category
event_type
component
component_id
location
message
event_id
```

Deduplicated normalized events only by default.

## Row 3 — Correlated telemetry

Prometheus panels scoped to server/component and dashboard time range.

Examples:

Storage:

```text
drive temperature
media errors
predictive failure
drive health
```

Memory:

```text
ECC correctable
ECC uncorrectable
DIMM health
```

Thermal:

```text
inlet temperature
sensor temperature
fan RPM
```

Use Grafana annotations from OpenSearch events where feasible.

## Drilldown

Event → Server Hardware Detail with:

```text
server
time range
component/location context
```

---

# Dashboard 04 — Hardware Inventory

## Purpose

Answer:

```text
What hardware is installed?
Where?
What firmware?
What changed?
What information do I give a vendor for RMA?
```

Datasource: OpenSearch primarily.

## Variables

```text
$server
$component
```

## Server identity

```text
vendor
model
service tag
BIOS
BMC firmware
```

## Component inventory

Tabs/rows:

```text
CPU
Memory
Storage
PSU
NIC
GPU
PCIe
```

Storage must expose:

```text
server
controller
enclosure
bay
manufacturer
model
serial
part number
firmware
capacity
health
```

Memory:

```text
server
slot
manufacturer
part number
serial
capacity
speed
health
```

## Firmware matrix

Show firmware by:

```text
server
component
location
model
firmware
```

Phase 2 can add drift highlighting.

## Replacement history

Phase 2 inventory-change documents.

Example:

```text
Bay 03
AAA → BBB
2026-08-29
```

## RMA view

An operator selecting a failed component should be able to copy:

```text
Server
Vendor
Server model
Service tag
Component
Physical location
Manufacturer
Model
Serial
Part number
Firmware
Capacity
Failure/event evidence
```

No need to implement automated vendor submission in MVP.

---

# Dashboard 05 — Incident Investigation

Phase 2.

Do not block Phase 1 on this dashboard.

Incident view correlates multiple events into one operational case:

```text
temperature high
→ media error
→ predictive failure
→ drive failure
→ RAID degraded
```

Display:

```text
incident id
server
severity
category
component
location
start/end
status
related event count
telemetry timeline
hardware identity
```

---

# Navigation Contract

```text
Dashboard 01 Overview
        ↓ server
Dashboard 02 Hardware Detail
     ├── event → Dashboard 03
     └── identity → Dashboard 04

Dashboard 03
     └── server/component → Dashboard 02

Dashboard 04
     └── failed component → Dashboard 02/03

Phase 2:
02/03/04 → Dashboard 05 Incident
```

Dashboard links must preserve `$server` and relevant time range.

# Grafana Provisioning Contract

Local Kind lab should provision:

```text
Prometheus datasource
OpenSearch datasource
Dashboard 01
Dashboard 02
Dashboard 03
Dashboard 04
```

Dashboard 01 is mandatory for Phase 1 visual validation.
02–04 may be delivered incrementally if their underlying data does not yet exist, but JSON/provisioning should follow this spec rather than inventing a different IA.

# Visual Design Rules

Operational density over decorative presentation.

Prefer:

```text
stat
table
state timeline
time series
heatmap/status grid where useful
```

Avoid:

```text
pie-chart-heavy overview
large decorative gauges
rainbow colors
panels with no troubleshooting action
```

Use consistent state semantics across all dashboards.

# Dashboard Acceptance Tests

Automated validation should at minimum verify dashboard provisioning files are loaded.

Human smoke validation should confirm:

1. Overview shows all mock servers.
2. Unreachable mock is clearly BMC OFFLINE, not hardware CRITICAL.
3. Partial mock shows affected resource error without making unrelated resources fail.
4. Slow/unreachable target shows increasing data age.
5. Clicking a server opens Server Hardware Detail.
6. Storage view can represent Bay 03 independently from serial identity.
7. Event view can filter by server/component/severity.
8. Inventory identity comes from OpenSearch rather than high-cardinality Prometheus labels.
