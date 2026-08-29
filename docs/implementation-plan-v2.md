# SalaryThief v2 Implementation Plan

## Guiding rule

Evolve the current codebase in small phases. Preserve working behavior unless it conflicts with an explicit v2 invariant.

# Phase 0 — Baseline and lab repair

Before feature changes:

- inspect current code paths
- run `go test ./...`
- run/build the current collector locally
- bring up the current Kind environment
- identify missing manifests or broken assumptions
- record baseline behavior in `docs/validation-results.md`

Deliverable: a reproducible baseline or a repaired local lab.

# Phase 1 — Background polling, cache, freshness, isolation

This is the first implementation target.

## Required behavior

Introduce or formalize:

```text
Scheduler
  ↓
Bounded worker pool
  ↓
Redfish collection
  ↓
Normalized cached state
  ↓
Prometheus registry/export
```

The HTTP `/metrics` handler must not make Redfish network calls.

## Suggested internal concepts

Use names appropriate to the existing codebase, but the design should contain equivalents of:

- target registry/config
- scheduler
- work item
- worker pool
- per-target state/cache
- collection result
- error classification
- freshness timestamps
- per-resource status

Avoid unnecessary abstraction if the prototype can be evolved cleanly.

## Acceptance criteria

Phase 1 is not complete until all of these are demonstrated:

- [ ] `go test ./...` passes.
- [ ] `/metrics` performs no synchronous BMC request.
- [ ] stopping one mock BMC does not make `/metrics` slow.
- [ ] one failed target does not stop updates from healthy targets.
- [ ] target connectivity can be represented as down independently of hardware health.
- [ ] last-success timestamp is exported.
- [ ] data age/staleness is exported.
- [ ] timeouts/auth/HTTP/schema failures can be distinguished internally and in self-observability where practical.
- [ ] concurrency is bounded.
- [ ] retry uses bounded policy/backoff rather than a tight loop.
- [ ] partial resource failure does not erase unrelated valid resource data.
- [ ] tests cover the cache/scrape decoupling behavior.
- [ ] local Kind smoke test passes.

# Phase 2 — Inventory model and change detection

Implement normalized physical topology and inventory storage.

Important fields include:

- server identity
- CPU socket/model/part/serial/cores/threads/health
- DIMM slot/manufacturer/part/serial/capacity/speed/health/ECC when available
- storage controller/enclosure/bay/model/serial/part/capacity/firmware/health
- PSU slot/model/serial/part/capacity/input/output/health
- fan location/model/speed/health
- NIC adapter/port/MAC/model/firmware/link
- GPU slot/model/serial/firmware/temp/power/health
- PCIe slot/device/vendor/device ID/firmware/health

Inventory refresh must have a slower independent cadence than telemetry.

Implement change detection with at least replacement semantics.

Acceptance highlights:

- inventory is not indexed every telemetry cycle
- Bay 03 AAA → BBB emits a replacement change
- serial/part/fw stay out of ordinary Prometheus labels

# Phase 3 — Event collection and deduplication

Use Redfish LogService/EventLog incremental polling for MVP.

Implement:

- cursor/incremental semantics appropriate to available Redfish data
- normalized severity/category/component/location
- dedup by stable event ID or fallback fingerprint
- bounded OpenSearch writer queue/retry behavior
- metrics for event ingest/dedup/errors

An OpenSearch outage must not break `/metrics`.

# Phase 4 — Grafana workflow

Build/revise dashboards around troubleshooting flow rather than a single all-purpose dashboard.

Target dashboards:

1. Physical Server Overview
2. Server Hardware Detail
3. Hardware Events & Timeline
4. Hardware Inventory

Variables should primarily be scoped by the selected datasource/environment and `$server`/`$component`/time range.

# Phase 5 — Incident correlation

Correlate related event sequences into incidents.

This is not required for the initial local architecture validation.

# Phase 6 — Vendor compatibility and production hardening

Add real hardware compatibility coverage and benchmark:

- Dell
- HPE
- Lenovo
- Supermicro

Track model family, firmware, standard resources, OEM resources, EventLog, EventService, and physical location quality.

Benchmark at 200/1k/2k/5k/10k logical targets and separately against real BMC hardware.

# Codex execution instructions

When this package is first introduced into a local checkout, use this sequence:

```text
1. Read AGENTS.md and referenced docs.
2. Inspect repository implementation.
3. Run tests.
4. Bring up existing Kind lab.
5. Fix only what is necessary to obtain a reproducible baseline.
6. Record baseline in docs/validation-results.md.
7. Implement Phase 1.
8. Add unit/integration tests.
9. Run tests.
10. Exercise healthy, unreachable, slow, and partial-failure mock scenarios.
11. Record exact commands and observed results.
12. Continue fixing until Phase 1 acceptance criteria pass.
```

Do not stop after generating code. Validation is part of the implementation task.
