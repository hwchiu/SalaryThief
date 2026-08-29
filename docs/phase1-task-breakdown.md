# Phase 1 Task Breakdown

This is an implementation-oriented decomposition for the first coding session.

## P1-01 — Baseline audit

Inspect current code and document:

- where target config is loaded
- where Redfish collection is initiated
- whether collection already runs in a goroutine/ticker
- how metrics are currently updated
- whether `/metrics` is only Prometheus handler state
- how child-resource errors are represented
- how OpenSearch writes are invoked
- current timeout/retry behavior

Do not change architecture until this is understood.

Output findings to `docs/validation-results.md`.

## P1-02 — Explicit collection state/cache

Introduce a concurrency-safe state holder if the existing metrics registry does not already provide the required semantics.

Per target, retain at minimum:

```text
last_attempt
last_success
connectivity
last_error_class
resource_family status
latest normalized telemetry
```

Requirements:

- reads for `/metrics` are local-memory only
- failed collection attempt does not automatically destroy the last known successful values
- freshness exposes that those values are old
- cache access is race-free under concurrent workers and Prometheus scrapes

Add race-oriented tests where practical.

## P1-03 — Scheduler

Create or formalize background scheduling.

Initial scope:

- telemetry/health cycle
- per-target due time
- bounded dispatch
- no tight retry loops

Do not implement a complex distributed scheduler.

Requirements:

- each target can fail independently
- scheduler keeps progressing when workers fail
- work is not duplicated uncontrollably
- shutdown respects context cancellation

## P1-04 — Bounded worker pool

Implement a fixed/configurable worker pool.

Tests must prove concurrency does not exceed configured worker count.

A test fake Redfish endpoint may block requests while tracking maximum in-flight requests.

## P1-05 — Error classification

Normalize at least:

```text
timeout
auth
http
tls
schema
network/unreachable
partial_resource
unknown
```

Do not leak password/token material into errors.

Connectivity and hardware health remain independent.

## P1-06 — Freshness metrics

Expose equivalents of:

```text
redfish_up
redfish_last_success_timestamp_seconds
redfish_data_age_seconds
```

If current metric names use the `bmc_` prefix, preserve compatibility when reasonable and document migration rather than performing an unnecessary breaking rename in Phase 1.

Data age definition:

```text
now - last_success
```

If no successful sample has ever existed, represent this explicitly and avoid pretending age is zero/fresh.

## P1-07 — Partial-resource behavior

A child resource error must not invalidate unrelated resource families.

Test example:

```text
Systems  200
Thermal  200
Storage  500
```

Expected:

```text
BMC connectivity = UP
system data       = current
thermal data      = current
storage           = ERROR/UNKNOWN
```

## P1-08 — Mock fault controls

Add deterministic controls for at least:

- healthy
- unreachable
- delayed response
- HTTP error for selected resource family

Prefer fixtures or small controllable proxy/mock services committed to the repo.

Avoid test behavior that depends on random timing.

## P1-09 — Integration smoke script

Create a repeatable command/script, e.g.:

```text
make integration-test
```

or an equivalent script under `hack/`.

It should fail non-zero when acceptance checks fail.

Minimum checks:

- collector ready
- `/metrics` responds
- healthy target is up
- failed target becomes down/stale
- healthy target keeps updating
- metrics endpoint remains fast during BMC failure

## P1-10 — Validation record

Codex must update `docs/validation-results.md` with:

- macOS architecture
- Go version
- Docker/runtime version
- Kind version
- kubectl version
- commands
- observed result
- failed attempts/fixes
- remaining limitations
