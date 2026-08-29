# Current Code Audit

This audit was prepared against the GitHub `main` branch before local Codex execution. Codex must verify every finding against the local checkout before modifying code.

## Executive summary

The current repository is not yet a reproducible baseline on `main`.

Two important findings change the Phase 1 plan:

1. The Prometheus HTTP scrape path already appears asynchronous from BMC polling.
2. The repository tree currently lacks several implementation/deployment files referenced by the existing code/scripts.

Therefore the first local task is **Phase 0 baseline repair**, followed by an incremental Phase 1.

## Repository baseline findings

Observed `main` tree contains:

```text
cmd/collector/main.go
configs/collector.yaml
configs/collector.kind.yaml
deploy/kind/00-namespace.yaml
deploy/kind/cluster.yaml
docs/*
hack/kind-up.sh
hack/kind-down.sh
Makefile
Dockerfile
go.mod
go.sum
```

However, `cmd/collector/main.go` imports:

```text
internal/collect
internal/config
internal/metrics
internal/opensearch
```

while the observed recursive Git tree did not contain `internal/`.

`hack/kind-up.sh` references:

```text
deploy/kind/10-redfish-mock.yaml
deploy/kind/20-collector.yaml
deploy/kind/30-opensearch.yaml
deploy/kind/40-prometheus.yaml
deploy/kind/50-grafana.yaml
deploy/kind/grafana-dashboard.json
```

while the observed recursive Git tree contained only:

```text
deploy/kind/00-namespace.yaml
deploy/kind/cluster.yaml
```

### Local verification requirement

Before reconstructing anything, Codex MUST run:

```bash
git status
git branch --show-current
git log -1 --oneline
find internal -maxdepth 3 -type f 2>/dev/null | sort
find deploy/kind -maxdepth 2 -type f | sort
go test ./...
```

The local checkout may contain unpushed files not visible through GitHub. Preserve them if present.

## cmd/collector/main.go audit

### What is already correct

`/metrics` is registered using the Prometheus handler:

```go
mux.Handle("/metrics", promhttp.Handler())
```

Redfish collection is invoked outside the HTTP handler by startup + ticker logic.

Therefore the core invariant:

```text
Prometheus scrape MUST NOT synchronously trigger BMC polling
```

appears to already be satisfied structurally.

Do NOT rewrite this behavior merely to satisfy the v2 document.

### Current scheduling model

Current conceptual behavior:

```text
ticker
  ↓
runOnce
  ↓
for every target:
    spawn goroutine
  ↓
wait for every target
  ↓
next cycle
```

Each target collection uses a context timeout and then updates the metric registry and optionally publishes to OpenSearch.

### Problems

#### Unbounded per-cycle concurrency

One goroutine is created per target.

At:

```text
200 targets  → up to ~200 concurrent target jobs
2,000 targets → up to ~2,000
10,000 targets → up to ~10,000
```

This is not an acceptable production concurrency policy.

#### Slowest-target barrier

`runOnce()` waits for every goroutine via a `WaitGroup`.

The next scheduler cycle is therefore coupled to the slowest collection job.

This produces collection drift when:

```text
cycle duration > configured scrape_interval
```

#### Single cadence

The current top-level scheduler appears to use one `ScrapeInterval` for collection.

v2 needs independent cadences for at least:

```text
telemetry
events
inventory
```

Inventory must not be published every telemetry cycle.

#### OpenSearch write in collection goroutine

The current top-level flow calls OpenSearch publishing inline after collection.

Consequences to verify:

- slow OpenSearch can consume collection worker time
- telemetry and persistence do not have clean backpressure isolation
- inventory/events may be emitted at telemetry cadence

This should be separated incrementally.

## Configuration audit

Production-like config currently provides:

```text
listen
scrape_interval
opensearch
targets[]
target timeout
target labels
collect resource switches
```

Kind config currently has two mock targets and OpenSearch enabled.

### Phase 1 config additions

Prefer backward-compatible additions such as:

```yaml
scheduler:
  telemetry_interval: 30s
  event_interval: 30s
  inventory_interval: 6h

workers:
  telemetry: 16
  events: 8
  inventory: 4

retry:
  initial_backoff: 5s
  max_backoff: 5m
```

Exact schema should fit the existing config implementation after it is recovered/verified.

Do not require all future knobs in the first commit.

## Dockerfile audit

The container build is simple and appropriate:

```text
Go 1.22 builder
CGO disabled
distroless non-root runtime
```

Preserve this unless implementation introduces a concrete requirement.

## Kind lab audit

The Kind cluster has host port mappings for:

```text
30000 Grafana
30090 Prometheus
30920 OpenSearch
```

This is a reasonable local UX and should be preserved.

The primary issue is completeness/reproducibility of the referenced manifests, not the basic cluster shape.

## Phase 0 conclusion

Before production architecture changes, local Codex should establish whether missing files are:

1. present locally but uncommitted,
2. present on another branch,
3. accidentally omitted from GitHub,
4. genuinely not implemented yet.

If they are genuinely missing, reconstruct the smallest correct baseline consistent with existing docs/config/scripts.

Do not fabricate history; document what was reconstructed.
