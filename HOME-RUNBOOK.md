# Home Handoff Runbook

This file is intentionally operational. The goal is to make the local handoff nearly mechanical.

## 1. Prepare repository

```bash
cd ~/path/to/SalaryThief
git status
git checkout -b codex/v2-phase1
```

Copy this handoff package into the repository root.

Do NOT overwrite locally newer project files except the handoff documents themselves.

## 2. Run the bootstrap check

From the repository root:

```bash
chmod +x BOOTSTRAP-CODEX.sh
./BOOTSTRAP-CODEX.sh
```

This only inspects the environment and runs the baseline Go test. It does not modify the project.

## 3. Optional host prerequisites

Verify:

```bash
go version
docker version
kind version
kubectl version --client
```

If Kind is missing and Homebrew is available:

```bash
brew install kind kubectl
```

Docker Desktop (or a compatible Docker API/runtime) should be running before `make kind-up`.

## 4. Start Codex from repo root

Use the prompt in `START-CODEX.md`.

Do not manually explain the architecture again unless Codex finds a contradiction in the local checkout.

## 5. What Codex should do first

Expected first actions:

```text
git status
inspect tree
read AGENTS.md
read referenced docs
go test ./...
inspect missing internal/*
inspect deploy/kind/*
attempt baseline build
attempt Kind lab
record failures
```

It should NOT immediately rewrite the collector.

## 6. Expected checkpoint

Before Phase 1 refactor, ask Codex to show:

```text
Baseline status:
- go test:
- docker build:
- make kind-up:
- collector /healthz:
- collector /metrics:
- Prometheus:
- OpenSearch:
- mock BMC targets:
- missing/reconstructed files:
```

If baseline is green, let it continue.

If baseline needed repair, review only the high-level reconstruction summary, then let it continue unless something looks wrong.

## 7. Phase 1 completion checkpoint

Ask Codex:

```text
Show me the Phase 1 acceptance table with PASS/FAIL and the exact command/evidence for every item. Do not mark anything PASS based only on code inspection.
```

Minimum required PASS:

```text
go test ./...
bounded workers
metrics scrape causes zero Redfish requests
failed BMC does not slow /metrics
healthy targets keep refreshing
last-success exported
data-age exported
unreachable classified
timeout classified
partial resource failure isolated
recovery demonstrated
Kind smoke test green
```

## 8. Final local review

Before commit:

```bash
git status
git diff --stat
git diff
go test ./...
```

Then optionally:

```bash
git add .
git commit -m "feat: add resilient background Redfish collection"
```

Once pushed to GitHub, ChatGPT can re-read the branch/PR and perform architecture review.
