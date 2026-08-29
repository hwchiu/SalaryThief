# Phase 1 Acceptance Evidence

Codex should fill this file with real results.

| Requirement | PASS/FAIL | Command / Evidence | Notes |
|---|---|---|---|
| `go test ./...` passes | PASS | Docker Go 1.22: `go test ./...` | Green (host Go 1.18 is unsupported by `go.mod`). |
| Docker image builds | PASS | `docker build -t salarythief-collector:dev .` | Green. |
| Kind lab starts | PASS | `make kind-up` | All required workload rollouts passed. |
| `/healthz` responds | PASS | `make integration-test` | Port-forwarded health probe passed. |
| `/metrics` responds | PASS | `make integration-test` | Repeated cache-only scrapes passed. |
| `/metrics` causes zero Redfish requests | PASS | `TestMetricsScrapeDoesNotCallRedfish` | Three metrics scrapes; counted zero mock requests. |
| bounded worker concurrency proven | PASS | `TestSchedulerBoundsConcurrency` | Eight targets, two-worker pool; observed maximum never exceeded two. |
| failed BMC does not block healthy targets | PASS | `make integration-test`; final scaled-mock probe | Unreachable/slow targets coexisted with `bmc-healthy-01` up/current; a total BMC failure exported resource health UNKNOWN (0). |
| failed BMC does not slow `/metrics` materially | PASS | `make integration-test` | `/metrics` succeeds under a 2s curl deadline during BMC faults. |
| `last_success` exported | PASS | `make integration-test` | Healthy target timestamp asserted. |
| `data_age` exported | PASS | `make integration-test` | Healthy target data-age asserted. |
| timeout classified | PASS | `TestScrapeClassifiesAuthAndTimeout` and slow lab target | Timeout class asserted; lab delay is 3s vs 1s timeout. |
| unreachable/network classified | PASS | `make integration-test` | Deterministic empty endpoint target exported down/stale state. |
| auth error classified | PASS | `TestScrapeClassifiesAuthAndTimeout` | 401 maps to `auth`. |
| HTTP error classified | PASS | `TestScrapeClassifiesPartialResourceFailure` | Storage HTTP 500 maps to per-resource error. |
| partial resource failure isolated | PASS | `make integration-test` | Partial target remains up; Storage metric equals ERROR (4). |
| recovery after fault demonstrated | PASS | `make integration-test` | Reattached endpoint returns target to up. |
| OpenSearch failure isolated from metrics | PASS | `make integration-test` | OpenSearch scaled to zero; cached metrics remained fast and healthy target remained up. |
| no sensitive credentials in metrics/logs | PASS | Metric labels are fixed to server/scope/resource/reason; `rg 'password|token' internal cmd` reviewed before final test | No credentials exposed by the collector metrics/log statements. |

## Evidence rules

A PASS requires an executed test/command or externally observable result.

Code inspection alone is insufficient for runtime behavior.
