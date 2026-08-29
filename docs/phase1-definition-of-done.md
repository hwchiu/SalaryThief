# Phase 1 Definition of Done

Phase 1 is complete only when implementation, testing, observability, and documentation all meet this definition.

## Build and unit quality

- `go test ./...` passes
- `go vet ./...` passes where applicable
- no obvious data races in concurrent state/worker tests
- Docker image builds
- shutdown is context-aware and clean

Recommended additional check:

```bash
go test -race ./...
```

If race mode is not feasible on the current environment, record why.

## Architecture

- `/metrics` contains no synchronous Redfish I/O
- background collection is explicit
- concurrency is bounded
- scheduling is not globally blocked by one slow target
- target failures are isolated
- partial-resource failures are isolated
- stale data semantics exist
- OpenSearch latency/failure does not block metrics serving

## Metrics

Evidence exists for:

- target up/down
- last successful collection
- data age
- collection/error counters
- bounded worker/queue visibility

No sensitive/high-cardinality hardware identity is added to ordinary telemetry labels.

## Failure scenarios

Demonstrate:

```text
healthy
unreachable
slow
partial resource failure
recovery
```

For every scenario, document exact commands and observed metrics/log behavior.

## Local integration

Kind lab starts from documented commands.

Minimum services:

```text
collector
Redfish mock(s)
Prometheus
OpenSearch
Grafana
```

If Grafana is not required for automated assertions, it must still start successfully if the baseline intends it to.

## Documentation

Update:

```text
docs/validation-results.md
docs/acceptance-evidence-template.md
```

If config/API/metrics changed, update relevant docs/examples.

## Reviewability

Final diff must be understandable as Phase 1.

Avoid unrelated formatting churn or broad renames.

Codex should provide:

```text
summary
files changed
architecture changes
tests added
commands executed
acceptance results
known limitations
```
