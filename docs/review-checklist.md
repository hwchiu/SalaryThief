# Post-Implementation Review Checklist

Use this after Codex says Phase 1 is complete.

## Architecture review

- Is any Redfish request reachable from the HTTP `/metrics` request path?
- Can one BMC timeout delay scheduling for unrelated targets?
- Is worker concurrency truly bounded?
- Is any queue unbounded?
- Can OpenSearch outage block telemetry collection?
- Are partial failures represented per resource family?

## State review

- What happens to last known data after a failed refresh?
- How is staleness exposed?
- Is never-successful state distinct from fresh zero?
- Are state transitions thread-safe?
- Does recovery clear the correct errors?

## Metrics review

- Any serial/part/fw/service-tag labels?
- Any raw URL/error-message labels?
- Are counter label values bounded?
- Is `redfish_up` strictly connectivity/data-access semantics?
- Are unknown/error states distinguishable from hardware critical?

## Scheduler review

- Is next work driven by due time rather than waiting for a whole global cycle?
- Does shutdown cancel pending work safely?
- Is backoff bounded?
- Could failing targets monopolize workers?

## Mock/test review

- Are faults deterministic?
- Is unreachable distinct from HTTP 500?
- Is partial failure actually resource-specific?
- Does `/metrics` isolation test count Redfish requests?
- Are tests polling conditions rather than relying on arbitrary sleeps?

## Security review

- Credentials absent from logs/metrics/OpenSearch fixtures?
- TLS verification defaults sane?
- Mock secrets obviously fake?
- No write/control BMC operations introduced?

## Diff review

Run:

```bash
git diff --stat
git diff
```

Look for:

- unrelated refactors
- deleted existing behavior
- generated junk
- secret material
- massive formatting-only changes
