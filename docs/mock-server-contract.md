# Mock Redfish Server Contract

The mock environment is a deterministic architecture test harness, not a vendor simulator.

## Required logical targets

Use stable names so scripts can assert against them:

```text
bmc-healthy-01
bmc-storage-fail-01
bmc-memory-fail-01
bmc-unreachable-01
bmc-slow-01
bmc-partial-01
```

Phase 1 only requires healthy, unreachable, slow, and partial targets to be fully automated. Storage/memory fault fixtures may be prepared for later phases.

## Fault injection

Faults should be controllable without rebuilding the collector.

Acceptable approaches:

- separate mock services/pods
- fixture-specific mock instances
- small reverse proxy with endpoint-specific delay/error rules
- mock API switch if deterministic and scriptable

Prefer Kubernetes-native controls that integration tests can automate.

## Slow target

Delay must exceed configured Redfish HTTP timeout by a comfortable deterministic margin.

Example:

```text
collector timeout: 2s
mock delay:        5s
```

The exact numbers may differ.

## Partial target

Only one resource family should fail.

Example:

```text
/redfish/v1                 200
/redfish/v1/Systems/...     200
.../Thermal                 200
.../Storage                 500
```

Do not simulate partial failure by killing the entire mock server.

## Unreachable target

The configured endpoint should become unreachable while remaining present in the target registry.

This scenario verifies:

- connectivity state
- cache freshness
- backoff
- isolation

## Assertions

The integration harness should assert behavior from externally observable surfaces where possible:

- collector metrics
- collector logs
- Prometheus query API
- OpenSearch API in later phases

Avoid tests that pass only by inspecting private implementation state.
