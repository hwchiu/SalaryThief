# Prometheus Metrics Contract

This document defines semantics that implementation must preserve. Metric names may remain backward-compatible with existing `bmc_*` metrics where practical, but semantics must not drift.

## Design rules

1. `/metrics` serves only local cached state.
2. BMC reachability is not hardware health.
3. Stale data must be distinguishable from current data.
4. High-cardinality identity belongs in inventory, not ordinary labels.
5. Missing data is not automatically zero.
6. A partial collection failure must affect only the relevant resource family.

## Common labels

Preferred bounded labels:

```text
observability_scope
server
component
component_id
location
```

Optional low-cardinality labels may include:

```text
vendor
role
site
rack
reason
resource_family
```

Avoid ordinary labels for:

```text
serial
part_number
firmware
service_tag
asset_tag
mac
raw_redfish_uri
event_message
```

## Connectivity / collection metrics

### `redfish_up`

Type: Gauge

Meaning:

```text
1 = collector can currently obtain the target's core Redfish data
0 = collector cannot currently obtain core Redfish data
```

It is NOT a hardware health metric.

Labels:

```text
observability_scope
server
```

### `redfish_last_success_timestamp_seconds`

Type: Gauge

Unix timestamp of last successful target-level collection.

Rules:

- update only after a successful core collection
- never set to `now` after a failed attempt
- if there has never been a success, either omit the series or expose a documented sentinel; prefer omission where implementation allows

### `redfish_data_age_seconds`

Type: Gauge

Definition:

```text
now - last_success
```

Rules:

- monotonically increases while no success occurs
- returns toward zero after a successful collection
- must not be reported as zero if no successful sample exists

### `redfish_collection_duration_seconds`

Type: Histogram or Gauge depending on current implementation.

Preferred: Histogram for request/collection latency analysis.

### `redfish_collection_errors_total`

Type: Counter

Labels:

```text
observability_scope
server
reason
```

`reason` is bounded to:

```text
timeout
auth
http
tls
network
schema
partial_resource
unknown
```

Do not put raw error text into labels.

## Collector self-observability

### `redfish_collector_targets`

Type: Gauge

Current number of configured targets.

### `redfish_collector_queue_depth`

Type: Gauge

Labels:

```text
queue
```

Expected bounded values:

```text
telemetry
events
inventory
persistence
```

### `redfish_collector_workers_active`

Type: Gauge

Labels:

```text
pool
```

### `redfish_collector_requests_total`

Type: Counter

Labels:

```text
resource_family
result
```

`result` examples:

```text
success
error
```

### `redfish_collector_request_duration_seconds`

Type: Histogram

Labels should remain bounded, preferably:

```text
resource_family
```

Do not label by full URL.

### `redfish_collector_request_errors_total`

Type: Counter

Labels:

```text
reason
resource_family
```

### `redfish_collector_persistence_errors_total`

Type: Counter

Labels:

```text
sink
reason
```

Expected sink:

```text
opensearch
```

## Hardware health metric semantics

Use a small numeric state encoding and document it once.

Recommended:

```text
0 = UNKNOWN
1 = OK
2 = WARNING
3 = CRITICAL
4 = ERROR
```

Do not use `ERROR` to mean hardware critical. `ERROR` means the collector could not reliably determine that resource's health.

Example:

```text
redfish_component_health{
  server="server-042",
  component="drive",
  component_id="drive-03",
  location="Bay 03"
} 3
```

## Telemetry metrics

Representative contract:

```text
redfish_temperature_celsius
redfish_fan_speed_rpm
redfish_inlet_temperature_celsius
redfish_power_watts
redfish_power_input_watts
redfish_power_output_watts
redfish_cpu_temperature_celsius
redfish_cpu_power_watts
redfish_memory_ecc_correctable_total
redfish_memory_ecc_uncorrectable_total
redfish_drive_temperature_celsius
redfish_drive_media_error_total
redfish_drive_predictive_failure
redfish_drive_health
```

### Missing data rule

If Redfish does not provide a value:

- do not synthesize `0` unless zero is semantically correct
- omit the series or retain last successful value with freshness semantics
- never turn "unknown" into "healthy"

## Partial-failure example

Suppose:

```text
Systems  = success
Thermal  = success
Storage  = failure
```

Expected:

```text
redfish_up = 1
system metrics = current
thermal metrics = current
storage health = ERROR/UNKNOWN
storage freshness indicates stale/error
```

The target must not become globally down merely because Storage failed.

## Backward compatibility

If the existing prototype exports `bmc_*` names, Phase 1 may:

1. keep them and add missing semantics,
2. add new `redfish_*` metrics in parallel,
3. defer a breaking rename.

A broad metric rename is not a Phase 1 requirement.
