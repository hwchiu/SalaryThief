# Phase 1 Canonical Test Sequence

Codex should run this sequence before declaring completion.

```text
A. Static/unit
B. Build
C. Baseline lab
D. Healthy
E. Unreachable
F. Recovery
G. Slow
H. Partial
I. OpenSearch outage
J. Final regression
```

## A

```bash
go test ./...
go vet ./...
go test -race ./...
```

Record unsupported checks rather than silently skipping them.

## B

```bash
docker build -t salarythief-collector:dev .
```

## C

```bash
make kind-up
kubectl get pods -A
```

## D–I

Prefer:

```bash
make integration-test
```

once automated.

## J

After all fixes:

```bash
go test ./...
make integration-test
git diff --check
git status
```

Then fill the acceptance evidence table.

No code change after the final test run may be included in a claimed PASS without rerunning relevant tests.
