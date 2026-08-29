# Grafana Query Contract

This document defines query intent. Codex must adapt exact metric names to the implemented Phase 1 compatibility surface.

## Prometheus examples

### Server list

```promql
redfish_up
```

Use label values of `server` for variables.

### Offline

```promql
redfish_up == 0
```

### Stale

Initial lab threshold:

```promql
redfish_data_age_seconds > 120
```

Production threshold must reflect collection cadence.

### Last success

```promql
redfish_last_success_timestamp_seconds
```

### Temperature

```promql
redfish_temperature_celsius{server="$server"}
```

### Fan

```promql
redfish_fan_speed_rpm{server="$server"}
```

### Power

```promql
redfish_power_watts{server="$server"}
```

### Component health

```promql
redfish_component_health{server="$server"}
```

### Drive health

```promql
redfish_drive_health{server="$server"}
```

### Drive predictive failure

```promql
redfish_drive_predictive_failure{server="$server"}
```

## OpenSearch query intent

Events filter:

```text
server:$server
AND severity:($severity)
AND component:($component)
```

Inventory filter:

```text
server:$server
```

Location filter when drilling down:

```text
server:$server AND location:"Bay 03"
```

## Datasource ownership

Prometheus:

```text
current health
connectivity
freshness
time-series telemetry
counters
```

OpenSearch:

```text
serial
part number
firmware
service tag
model/manufacturer identity
raw/normalized events
inventory history
replacement history
```

Do not duplicate hardware identity into Prometheus merely because Grafana joins are inconvenient.

## Joins

Where Grafana needs both live status and inventory identity, prefer dashboard/table transformations or separate panels.

Do not corrupt backend data models to simplify a visualization.
