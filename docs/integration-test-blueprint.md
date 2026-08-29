# Integration Test Blueprint

## Entry point

Target developer experience:

```bash
make integration-test
```

Suggested flow:

```text
preflight
→ ensure lab
→ wait readiness
→ baseline assertions
→ fault scenarios
→ recovery
→ report
```

## Script structure

Recommended:

```text
hack/
├── integration-test.sh
└── lib/
    └── test-helpers.sh
```

Do not require Bash features unavailable in macOS default shell unless documented.

## Helper primitives

Implement equivalents of:

```text
wait_for_condition
prom_query
metric_value
assert_eq
assert_gt
assert_lt
kubectl_wait
dump_diagnostics
```

On failure, print actionable diagnostics.

## Baseline assertions

1. collector health endpoint succeeds
2. collector metrics endpoint succeeds
3. Prometheus is ready
4. OpenSearch is ready
5. healthy target becomes `up`
6. last-success exists
7. data age is bounded after successful collection

## Scrape-isolation assertion

Measure `/metrics` while a BMC is unavailable/slow.

Use a generous threshold appropriate for a laptop.

More importantly, prove Redfish request count does not increase as a consequence of additional Prometheus/HTTP scrapes.

## Unreachable scenario

Action:

```text
make bmc-unreachable-01 unavailable
```

Assert:

```text
target up → 0
data age increases
healthy target remains up
healthy target last-success continues advancing
```

Then restore and assert recovery.

## Slow scenario

Action:

```text
enable deterministic delay > collector timeout
```

Assert:

```text
timeout error counter increases
worker count remains bounded
metrics endpoint remains responsive
other targets refresh
```

## Partial scenario

Action:

```text
Storage → HTTP failure
other resource families → success
```

Assert:

```text
target connectivity remains up
storage status becomes error/unknown
thermal/system remain current
```

## OpenSearch isolation scenario

Temporarily make OpenSearch unavailable.

Assert:

```text
metrics continue to serve
healthy BMC collection continues
persistence error/queue signal appears
```

Restore OpenSearch and verify the collector remains healthy.

## Exit behavior

Any failed assertion:

```text
dump diagnostics
exit non-zero
```

Success:

```text
print concise PASS table
exit 0
```

## CI compatibility

Even if initial execution is local macOS only, scripts should avoid assumptions that prevent later use in GitHub Actions or another Linux CI runner.
