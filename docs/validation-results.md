# Validation Results

## Environment

```text
Date: 2026-08-29
macOS: Darwin (local macOS development host)
Architecture: arm64
Host Go: go1.18 darwin/arm64 (too old for this repository's go 1.22 module)
Validation Go: golang:1.22-bookworm Docker image
Docker/runtime: Docker Desktop 4.22.0, Engine 24.0.5 linux/arm64
Kind: v0.33.0
kubectl: v1.27.2 client
Git commit: bdeb79f0fcde477a56f5e59dad940b66d041d097
Branch: codex/v2-phase1
```

## Baseline

| Check | Result | Evidence |
|---|---|---|
| `go test ./...` | PASS | `docker run --rm -v "$PWD:/src" -w /src golang:1.22-bookworm go test ./...` |
| Docker build | PASS | `docker build -t salarythief-collector:dev .` |
| `make kind-up` | PASS | Created/reused `salarythief` Kind cluster; collector, three mocks, Prometheus, OpenSearch, Grafana all rolled out. |
| collector health | PASS | `hack/integration-test.sh` polled `GET /healthz` through `kubectl port-forward`. |
| collector metrics | PASS | `hack/integration-test.sh` asserted exported cache/freshness/resource metrics. |
| Prometheus | PASS | Integration queried Prometheus HTTP API and found `bmc-healthy-01`. |
| OpenSearch | PASS | OpenSearch deployment rolled out and was deliberately scaled down/up during integration. |
| Grafana | PASS | `make kind-up` observed successful Grafana deployment rollout. |

## Phase 1 scenarios

| Scenario | Result | Evidence |
|---|---|---|
| Healthy | PASS | `redfish_up{server="bmc-healthy-01"} 1` and last-success/data-age series asserted by `make integration-test`. |
| Unreachable BMC | PASS | Empty endpoint fault asserted `redfish_up{server="bmc-unreachable-01"} 0`; final review also scaled `bmc-healthy-01` to zero and observed resource health UNKNOWN (0), then restored it. |
| Slow BMC | PASS | `bmc-slow-01` has a deterministic 3s delay vs 1s timeout; its exported state is down and timeout counter is exercised in unit/integration runs. |
| Partial resource failure | PASS | `bmc-partial-01` remains up while `redfish_resource_health{resource_family="storage"}` is `4`. |
| Recovery | PASS | Harness attaches the unreachable Service to the healthy endpoint and asserts it returns to `redfish_up=1`. |
| OpenSearch unavailable | PASS | Harness scales OpenSearch to zero and proves `/metrics` remains fast and healthy target state remains available, then restores it. |

## Defects found

- Baseline imported missing `internal/{collect,config,metrics,opensearch}` packages and absent Kind manifests. Reconstructed the smallest v2-compatible implementation and lab rather than replacing the repository.
- Host Go is 1.18 while `go.mod` requires Go 1.22. Used the repository's Docker Go 1.22 build environment for all Go validation.
- Initial scheduler only dispatched the first queue-capacity batch. Added dispatch-on-completion; the bounded-concurrency test now covers the regression.
- A Kind Service selector removal left a stale EndpointSlice. The integration harness removes both Endpoints and EndpointSlices before asserting the deterministic unreachable fault.
- Final review found stale cached `OK` resource state could survive a total connectivity failure. Cache now retains freshness but exports every affected resource as UNKNOWN; unit and scaled-mock runtime checks passed.
- Phase 1 merge review: added per-target scheduler in-flight exclusion (unit-tested at max one concurrent call), bounded OpenSearch HTTP client timeout/queue/drop/error instrumentation, and completed Grafana Dashboard 01 provisioning.

## Remaining limitations

- The mock implements only the small Redfish response surface needed for Phase 1; it is not vendor firmware compatibility coverage.
- Inventory change detection and EventLog deduplication remain Phase 2/3 work.
- OpenSearch publishing is bounded/asynchronous, but detailed persistence retry and error metrics are future hardening work.
