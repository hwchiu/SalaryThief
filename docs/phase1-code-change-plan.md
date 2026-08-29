# Phase 1 Code Change Plan

Codex must verify file paths after recovering the local baseline. Names below are architectural targets, not mandatory filenames.

## Stage A — Recover baseline

Goal:

```text
go test ./...
docker build .
make kind-up
```

must become reproducible before major refactoring.

Tasks:

- recover/implement missing `internal/*`
- recover/implement referenced Kind manifests
- confirm two healthy mock targets work
- confirm Prometheus scrapes collector
- confirm OpenSearch receives current prototype documents
- record exact baseline

Do not optimize yet.

## Stage B — Preserve scrape decoupling

Add a regression test proving HTTP `/metrics` does not issue Redfish requests.

A suitable test design:

1. create a fake Redfish server with request counter
2. populate collector state once, or initialize metric state directly
3. record request count
4. perform repeated `/metrics` scrapes
5. assert Redfish request count did not increase

This protects the behavior that already appears correct.

## Stage C — Replace fan-out-per-cycle with bounded workers

Current:

```text
for targets:
  go scrape(target)
wait all
```

Target:

```text
scheduler
    ↓
bounded queue
    ↓
N workers
    ↓
collect target
```

Required properties:

- configurable upper concurrency bound
- context-aware shutdown
- no goroutine leak
- queue depth observable
- one failing target cannot stop worker pool
- no global wait barrier before future target scheduling

Initial implementation may use one telemetry pool.

Do not implement three fully generic worker frameworks if one clean reusable primitive suffices.

## Stage D — Per-target collection state

Introduce a concurrency-safe target state abstraction.

Minimum semantic fields:

```text
target/server ID
last_attempt
last_success
connectivity
last_error_class
data/resource freshness
latest successful snapshot/resource values
```

Rules:

- collection attempt updates `last_attempt`
- successful relevant collection updates `last_success`
- total connectivity failure sets connectivity down
- total failure does not pretend hardware health is critical
- stale values can remain exposed if clearly marked stale
- partial resource error changes only affected resource status

## Stage E — Freshness/self metrics

Implement or map existing names to:

```text
redfish_up
redfish_last_success_timestamp_seconds
redfish_data_age_seconds
```

Add collector internals:

```text
queue depth
active workers
request errors by reason
timeouts
```

Avoid a broad breaking metric rename in Phase 1.

If current `bmc_*` names exist, compatibility aliases are acceptable.

## Stage F — Error classification

Create a small stable enum/type.

Suggested values:

```go
type ErrorClass string

const (
    ErrorNone      ErrorClass = ""
    ErrorTimeout   ErrorClass = "timeout"
    ErrorAuth      ErrorClass = "auth"
    ErrorHTTP      ErrorClass = "http"
    ErrorTLS       ErrorClass = "tls"
    ErrorNetwork   ErrorClass = "network"
    ErrorSchema    ErrorClass = "schema"
    ErrorPartial   ErrorClass = "partial_resource"
    ErrorUnknown   ErrorClass = "unknown"
)
```

Exact names may differ.

The important part is deterministic classification and test coverage.

## Stage G — Split OpenSearch persistence from telemetry execution

Do not let an OpenSearch outage consume or block the metrics-serving path.

Minimum Phase 1 option:

```text
collector result
  ├── metrics/cache update immediately
  └── bounded persistence queue → OpenSearch worker
```

If full event/inventory separation is too large for Phase 1, at minimum:

- isolate OpenSearch write latency
- expose queue/write failures
- use bounded buffering
- define explicit drop/retry behavior

Never use an unbounded channel.

## Stage H — Independent cadences

At minimum prepare separate configuration/state for:

```text
telemetry interval
inventory interval
events interval
```

Phase 1 may fully implement telemetry scheduling and only scaffold later schedules if baseline code does not yet have separate collectors.

But it MUST stop treating full inventory publication as an unavoidable every-telemetry-cycle operation.

## Stage I — Fault-capable mock lab

Add these deterministic targets:

```text
bmc-healthy-01
bmc-unreachable-01
bmc-slow-01
bmc-partial-01
```

The test harness must control them.

Preferred implementation pattern:

```text
mock fixture/service
        +
fault proxy or purpose-built small mock service
```

Choose the least complex approach compatible with the current DMTF mock setup.

## Stage J — Automated integration validation

Add an entry point such as:

```bash
make integration-test
```

It should:

```text
start/verify lab
wait readiness
assert healthy target
trigger unreachable fault
assert redfish_up=0
measure /metrics latency
restore target
assert recovery
trigger slow fault
assert timeout classification/backoff
trigger partial fault
assert unaffected families remain current
exit non-zero on failure
```

Tests should use generous deterministic thresholds suitable for macOS laptops.

Do not use fragile millisecond-level assertions.

## Recommended commit sequence

Keep changes reviewable:

```text
1. chore: restore reproducible local baseline
2. test: protect asynchronous metrics scrape behavior
3. feat: add target state and freshness
4. feat: add bounded collection workers
5. feat: classify collection failures
6. feat: isolate persistence path
7. test: add Redfish failure scenarios
8. test: add local integration validation
9. docs: record validation results
```

Codex may combine commits locally, but should keep conceptual separation in the diff.
