# Developer Command Contract

The repository should converge on these commands.

```bash
make test
make build
make kind-up
make kind-down
make integration-test
```

Optional useful commands:

```bash
make kind-status
make kind-logs
make test-race
```

## Semantics

### `make test`

Run fast Go unit/component tests.

### `make build`

Build collector locally and/or container image according to current project convention.

### `make kind-up`

Create/reuse Kind cluster, build/load current collector image, deploy mocks and observability stack, and wait for essential readiness.

Must be idempotent enough for normal iterative development.

### `make kind-down`

Delete the SalaryThief Kind cluster safely.

### `make integration-test`

Run deterministic Phase 1 acceptance scenarios against the local lab.

It must return non-zero on failure.

### `make kind-status`

Optional concise:

```text
kubectl get pods,svc -A
```

### `make kind-logs`

Optional collector/mock diagnostics.

## Rule

Do not hide failing commands simply to keep Make targets green.
