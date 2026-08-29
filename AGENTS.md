# SalaryThief Agent Instructions

You are working on SalaryThief, a Go-based Redfish/BMC observability collector.

Before changing code, read these documents in order:

1. `docs/chatgpt-design-handoff.md`
2. `docs/architecture-v2.md`
3. `docs/local-validation.md`
4. `docs/implementation-plan-v2.md`
5. `docs/phase1-task-breakdown.md`
6. `docs/test-matrix.md`
7. `docs/mock-server-contract.md`
8. `docs/architecture-decisions.md`
9. `docs/current-code-audit.md`
10. `docs/phase1-code-change-plan.md`
11. `HOME-RUNBOOK.md`
12. `docs/metrics-contract.md`
13. `docs/domain-model.md`
14. `docs/phase2-contract.md`
15. `docs/acceptance-evidence-template.md`
16. `docs/agent-guardrails.md`
17. `docs/phase1-definition-of-done.md`
18. `docs/review-checklist.md`
19. `docs/recovery-and-rollback.md`
20. `docs/kind-lab-blueprint.md`
21. `docs/integration-test-blueprint.md`
22. `docs/makefile-contract.md`
23. `docs/phase1-test-sequence.md`
24. `docs/grafana-dashboard-spec.md`
25. `docs/grafana-query-contract.md`
26. `docs/grafana-mock-acceptance.md`

Then inspect the existing implementation, especially:

- `cmd/collector/`
- `internal/redfish/`
- `internal/collect/`
- `internal/metrics/`
- `internal/opensearch/`
- `configs/`
- `deploy/kind/`
- `hack/kind-up.sh`
- `Makefile`

## Primary objective

Evolve the existing prototype into the v2 architecture incrementally. Do not rewrite the repository from scratch.

The local macOS development environment is the primary concept-validation environment. The expected workflow is:

```text
inspect existing code
      ↓
create/repair local Kind + mock Redfish lab
      ↓
implement one small phase
      ↓
run unit tests
      ↓
run integration tests
      ↓
exercise failure scenarios
      ↓
inspect Prometheus/OpenSearch/Grafana behavior
      ↓
fix defects
      ↓
repeat until acceptance criteria pass
```

## Architecture invariants

These are requirements, not suggestions.

1. `/metrics` MUST NOT synchronously call any BMC.
2. BMC polling is owned by background scheduler/worker logic.
3. Prometheus reads already-collected state from an in-memory/cache layer.
4. A slow or unreachable BMC must not make `/metrics` slow.
5. One failed BMC must not block collection from other BMCs.
6. Collection concurrency must be bounded.
7. Telemetry, events, and inventory must have independent scheduling and failure domains.
8. Inventory must not be written every telemetry cycle.
9. Prometheus stores low-cardinality telemetry/health signals.
10. OpenSearch/Elasticsearch stores inventory, event documents, inventory changes, and later incidents.
11. Serial number, part number, firmware version, service tag, asset tag, and similar identity fields must not become normal Prometheus labels.
12. Physical location and hardware identity are separate concepts.
13. Vendor-specific Redfish differences must terminate at an adapter/normalization boundary.
14. Partial collection failure must be visible as UNKNOWN/ERROR/stale without destroying previously valid unrelated data.
15. No BMC credentials, session tokens, or sensitive configuration may be written to metrics, logs, OpenSearch, fixtures, or Git.
16. The collector remains read-only with respect to server power/control/boot mutation/virtual media unless explicitly changed by a future design decision.

## Work style

- Start by inspecting the current implementation and identifying what already satisfies v2.
- Prefer minimal, reviewable changes.
- Add or update tests with every behavior change.
- Run `go test ./...` frequently.
- Use the existing `make test`, `make kind-up`, and `make kind-down` workflow when practical.
- If existing Kind manifests referenced by `hack/kind-up.sh` are missing or inconsistent, repair them as part of the local validation environment.
- Do not claim a scenario is validated unless you actually ran the corresponding test or command and observed the result.
- Keep a concise implementation/validation log in `docs/validation-results.md` while working.

## First task

Do NOT immediately redesign everything.

First:

1. Inspect the current repository.
2. Run the existing unit tests.
3. Inspect whether the existing Kind lab is complete and runnable on macOS with Docker Desktop or another Docker-compatible runtime.
4. Bring the baseline local lab up if possible.
5. Record baseline failures.
6. Produce a short implementation plan for Phase 1.
7. Then implement Phase 1 and validate it.

Phase 1 is complete only when all acceptance criteria in `docs/implementation-plan-v2.md` pass.
