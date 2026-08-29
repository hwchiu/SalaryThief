# Kind Lab Implementation Blueprint

## Objective

Provide a deterministic local macOS lab that exercises SalaryThief without physical BMCs.

```text
macOS
  └─ Kind cluster
      ├─ salarythief namespace
      │   ├─ collector
      │   ├─ redfish healthy
      │   ├─ redfish slow
      │   └─ redfish partial
      └─ observability namespace
          ├─ Prometheus
          ├─ OpenSearch
          └─ Grafana
```

An unreachable target may be represented by a configured Service/endpoint with no ready backing endpoint, or by scaling a dedicated mock deployment to zero.

## Preserve existing UX

Existing host mappings should remain:

```text
Grafana    localhost:30000
Prometheus localhost:30090
OpenSearch localhost:30920
```

Collector may continue to use port-forwarding unless a stable host mapping is useful.

## Recommended manifests

If the files referenced by the existing `kind-up.sh` are genuinely missing, reconstruct:

```text
deploy/kind/
├── 00-namespace.yaml
├── 10-redfish-mock.yaml
├── 15-redfish-faults.yaml
├── 20-collector.yaml
├── 30-opensearch.yaml
├── 40-prometheus.yaml
├── 50-grafana.yaml
├── grafana-dashboard.json
└── cluster.yaml
```

Exact decomposition may differ if the local checkout contains newer files.

## Mock design

Prefer DMTF Redfish mock fixtures for representative payloads.

For deterministic network faults, a small fault proxy/service is acceptable.

Required behaviors:

```text
healthy:
  normal Redfish responses

slow:
  delay > collector timeout

partial:
  normal ServiceRoot/System/Thermal
  Storage returns deterministic HTTP error

unreachable:
  target configured but endpoint unavailable
```

Avoid relying on random packet loss.

## Collector configuration

Kind config should use stable logical target names:

```text
bmc-healthy-01
bmc-unreachable-01
bmc-slow-01
bmc-partial-01
```

Use fake local credentials only.

## Prometheus

Prometheus scrapes SalaryThief only.

Prometheus must NOT scrape Redfish mock services.

Recommended scrape interval for lab:

```text
5s–15s
```

This keeps tests fast while remaining deterministic.

## OpenSearch

Use a single-node development configuration with security disabled only for the local Kind lab.

Do not copy local insecure settings into production config.

## Grafana

Provision:

```text
Prometheus datasource
OpenSearch datasource
at least one SalaryThief overview dashboard
```

Grafana is useful for human inspection but automated acceptance tests should use APIs/metrics rather than screenshots.

## Readiness

`kind-up.sh` should eventually fail if essential components cannot become ready.

Avoid masking rollout failures with unconditional `|| true` in the final reliable workflow.

Provide diagnostics before exit:

```text
kubectl get pods -A
kubectl get svc -A
kubectl describe ...
kubectl logs ...
```

when startup fails.

## Cleanup

`make kind-down` must be idempotent.

A failed test run should not require manual Kubernetes object cleanup before the next attempt.
