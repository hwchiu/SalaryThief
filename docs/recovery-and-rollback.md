# Recovery and Rollback Guide

## If baseline repair becomes messy

Keep baseline restoration separate from Phase 1 changes.

Recommended commits:

```text
chore: restore reproducible local baseline
feat: add resilient background collection
```

This makes rollback straightforward.

## If Phase 1 breaks the local lab

Use Git to inspect, not destructively reset:

```bash
git status
git diff
git log --oneline --decorate -10
```

Revert only the agent's own commits if needed.

Do not destroy pre-existing user changes.

## If a new scheduler causes regressions

Fallback strategy:

1. retain existing ticker behavior behind a compatibility path,
2. isolate worker-pool changes,
3. add tests proving equivalence for healthy targets,
4. switch default only after fault tests pass.

Avoid a large one-way migration without tests.

## If OpenSearch isolation is unstable

Metrics correctness has priority.

Temporarily disable persistence in local config and prove:

```text
collector
cache
metrics
Prometheus
```

remain healthy.

Then repair persistence separately.

## If Kind is the blocker

Separate host/runtime issue from application issue.

Record:

```text
docker info
kind version
kubectl version --client
kind create cluster ...
```

Do not alter application architecture to work around a broken local container runtime.
