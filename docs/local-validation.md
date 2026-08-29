# Local macOS Validation Plan

## Goal

The local environment must allow a developer or Codex to validate core architecture and behavior without access to a heterogeneous physical server lab.

It validates protocol/data-flow contracts and failure handling. It does NOT certify vendor firmware compatibility.

## Expected host prerequisites

Prefer tooling that works on macOS:

- Git
- Go (matching `go.mod`)
- Docker Desktop or a Docker-compatible runtime usable by Kind
- `kind`
- `kubectl`
- `curl`
- optionally `jq`

The agent should detect missing dependencies and report exact commands/prerequisites rather than silently skipping tests.

## Existing entry points

Use the repository's existing workflow when possible:

```bash
make test
make kind-up
make kind-down
```

The current Kind bootstrap is expected to build the collector image, create/load it into Kind, apply mock BMC and observability manifests, and expose Grafana/Prometheus/OpenSearch.

If referenced manifests are missing or broken, repair the lab rather than replacing it with an unrelated environment.

## Mock profiles

Create enough deterministic Redfish mock profiles to exercise these cases.

### A. Healthy server

- BMC reachable
- System OK
- CPU OK
- memory OK
- storage OK
- temperatures normal
- PSU/fan OK

### B. Storage degradation

- Bay 03 Warning/Critical
- predictive failure
- storage EventLog entry
- serial/location known

Expected:

- drive health metric changes
- event document appears once
- physical location is `Bay 03`
- inventory still identifies the exact drive

### C. DIMM problem

- DIMM A2 Warning/Critical
- ECC signal/event when model supports it

Expected:

- memory health is degraded
- DIMM A2 is identifiable
- unrelated storage remains healthy

### D. BMC unreachable

Stop or isolate one mock endpoint.

Expected:

- `redfish_up=0`
- hardware state becomes UNKNOWN/stale, not falsely CRITICAL
- `redfish_data_age_seconds` increases
- other targets continue updating
- `/metrics` stays fast

### E. Slow BMC

Add controllable delay greater than the configured Redfish timeout.

Expected:

- timeout is classified correctly
- retries/backoff occur according to policy
- bounded worker pool prevents runaway goroutines/connections
- `/metrics` latency is unaffected

### F. Partial Redfish failure

Make Systems/Thermal valid but Storage fail with HTTP/schema error.

Expected:

```text
BMC      UP
System   valid
Thermal  valid
Storage  UNKNOWN/ERROR
```

Previously valid unrelated telemetry must remain available.

### G. Hardware replacement

Initial inventory:

```text
Bay 03 → Serial AAA
```

Then mutate fixture/server response:

```text
Bay 03 → Serial BBB
```

Expected:

- location remains Bay 03
- new hardware identity is stored
- inventory change `replaced` is emitted
- old/new serials are available in the change document, not Prometheus labels

### H. Duplicate EventLog polling

Return the same event on multiple collection cycles.

Expected:

- normalized event is indexed once
- dedup counter increments for duplicates

## Fleet sizes

The lab should support generating logical target fleets at least at:

```text
20
200
1,000
2,000
5,000
```

Use duplicated mock profiles with unique logical server IDs/URLs where practical.

Measure:

- collector CPU/memory
- BMC request rate
- active workers
- queue depth
- collection lag
- data age
- `/metrics` response time
- OpenSearch ingestion failures/retries

## Minimum smoke validation

The agent must automate or document commands that prove:

1. `go test ./...` passes.
2. Kind lab starts successfully.
3. collector `/metrics` is reachable.
4. Prometheus has discovered the collector.
5. OpenSearch is reachable.
6. at least one healthy Redfish target produces metrics.
7. stopping a mock BMC changes only that target's connectivity/freshness state.
8. `/metrics` remains responsive while the failed target is timing out/backing off.
9. at least one partial-resource failure is represented correctly.
10. at least one duplicate event is deduplicated.
11. at least one inventory replacement is detected.

## What this lab does not prove

Do not claim validation of:

- Dell iDRAC firmware quirks
- HPE iLO firmware quirks
- Lenovo XCC firmware quirks
- Supermicro BMC firmware quirks
- real EventService reliability
- real-world BMC rate limits
- TLS/cipher oddities
- OEM location naming consistency
- vendor-specific schema deviations not represented in fixtures

A later compatibility matrix should track real hardware model + firmware coverage separately.
