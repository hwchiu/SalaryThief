# Grafana Mock Acceptance Scenarios

Use the local mock fleet to make dashboard behavior visually obvious.

| Target | Expected Overview |
|---|---|
| bmc-healthy-01 | BMC UP, hardware OK |
| bmc-unreachable-01 | BMC OFFLINE, hardware UNKNOWN/stale |
| bmc-slow-01 | timeout/stale behavior without UI freeze |
| bmc-partial-01 | BMC UP, Storage ERROR/UNKNOWN, unrelated resources OK |

## Storage fixture

At least one mock should expose:

```text
Controller 0
Enclosure 1
Bay 01 → Serial DRIVE001 → OK
Bay 02 → Serial DRIVE002 → OK
Bay 03 → Serial DRIVE003 → CRITICAL/predictive failure
Bay 04 → Serial DRIVE004 → OK
```

This makes Dashboard 02 immediately useful.

## Memory fixture

At least one later fixture should expose:

```text
DIMM A1 OK
DIMM A2 CRITICAL
DIMM B1 OK
DIMM B2 OK
```

## Replacement fixture

Phase 2:

```text
Bay 03:
DRIVE003 → DRIVE103
```

The dashboard must keep Bay 03 constant while identity changes.

## Event fixture

Representative events:

```text
drive temperature warning
media error
predictive failure
drive failure
DIMM ECC uncorrectable
PSU warning
```

This enables Dashboard 03 development without real hardware.
