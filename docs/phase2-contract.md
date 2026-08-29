# Phase 2 Compatibility Contract

Phase 1 implementation must avoid choices that make Phase 2 inventory/events unnecessarily difficult.

## Inventory cadence

Inventory is slow-changing.

Default target:

```text
6h
```

Additional triggers:

```text
startup
manual refresh
hardware-change event
BMC/server reboot
inventory mismatch
```

Do not couple inventory writes to every telemetry collection.

## Inventory document identity

Preferred stable document keys:

```text
observability_scope
server
observed_at
```

Component identity must include:

```text
component
component_id
location
serial
part_number
firmware
```

## Replacement detection

For a physical location:

```text
old location = Bay 03
old serial   = AAA

new location = Bay 03
new serial   = BBB
```

emit:

```json
{
  "event_type": "inventory_change",
  "change": "replaced",
  "location": "Bay 03",
  "old_serial": "AAA",
  "new_serial": "BBB"
}
```

Phase 1 cache/domain model should therefore preserve both location and hardware identity separately.

## Event dedup compatibility

Future event polling should support:

Preferred:

```text
server + source + event_id
```

Fallback:

```text
server + component + location + event_type + message
```

Phase 1 target/component identifiers should be stable enough to use in these fingerprints.

## OpenSearch isolation

Phase 1 persistence queue should be designed so Phase 2 can add document types:

```text
events
inventory
inventory_change
```

without rebuilding the entire scheduling model.
