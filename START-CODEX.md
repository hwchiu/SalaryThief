# Start Codex Here

After copying this handoff package into the root of your local `SalaryThief` checkout, start Codex from that repository and give it this prompt:

```text
Read AGENTS.md and every design document it references.

You are operating directly in my local SalaryThief repository on macOS.
Your task is to continue the project, not just provide advice.

First inspect the existing implementation and run the current tests.
Then inspect and bring up the existing Kind + Redfish mock + Prometheus + OpenSearch + Grafana test environment.
If the current local environment is incomplete or broken, repair it while preserving the repository's intended design.

After obtaining a reproducible baseline, implement Phase 1 from docs/implementation-plan-v2.md.

You must actually run tests and validation commands after implementation. Exercise the healthy, BMC-unreachable, slow-BMC, and partial-Redfish-failure scenarios described in docs/local-validation.md. Fix issues you discover and rerun the tests.

Maintain docs/validation-results.md containing:
- environment/tool versions
- commands executed
- test results
- scenarios validated
- failures found and fixes made
- any remaining limitations

Do not stop after writing code. Continue until Phase 1 acceptance criteria pass, or until you hit a concrete external blocker that cannot be fixed from the repository. If blocked, document the exact blocker and all completed work.
```

## Suggested first local commands

Before starting Codex, from your SalaryThief checkout:

```bash
git checkout -b codex/v2-phase1

# copy this package's AGENTS.md, START-CODEX.md and docs/* into the repo

git status
```

Then launch Codex from the repository root.


## Important pre-audit finding

The GitHub `main` branch observed before this handoff appeared incomplete:
- `cmd/collector/main.go` imports several `internal/*` packages that were not present in the observed Git tree.
- `hack/kind-up.sh` references several `deploy/kind/*` manifests that were not present in the observed Git tree.

Your local checkout is authoritative. Before creating replacement files, verify whether they exist locally, are untracked, or exist on another local branch.

Also note: the current `main.go` already appears to keep `/metrics` independent from BMC requests. Preserve that property and add a regression test instead of unnecessarily rewriting the HTTP scrape path.


## Additional execution requirements

Before implementing Phase 1, read:

- `docs/metrics-contract.md`
- `docs/domain-model.md`
- `docs/phase2-contract.md`

Treat these as semantic contracts, but adapt implementation to the existing repository rather than forcing exact type/file names.

During validation, fill:

- `docs/validation-results.md`
- `docs/acceptance-evidence-template.md`

A runtime requirement may be marked PASS only if it was actually executed/observed.


## Guardrails and completion

Before changing code, read `docs/agent-guardrails.md`.

Use `docs/phase1-definition-of-done.md` as the final completion gate.

When you believe Phase 1 is done, perform the checks in `docs/review-checklist.md` and fill both validation documents with real evidence.

If you hit a stop condition, do not improvise around it. Document the blocker using the required format.


## Grafana

Read the Grafana specifications before changing dashboard JSON or datasource provisioning.

For Phase 1, Dashboard 01 (Physical Server Overview) is the mandatory visual dashboard. Preserve the information architecture for Dashboards 02–04 and implement them as underlying telemetry/inventory/event data becomes available. Do not invent a separate dashboard architecture.
