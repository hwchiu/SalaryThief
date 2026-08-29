# Agent Guardrails

These guardrails exist to prevent unnecessary rewrites, hidden failures, and unreviewable changes.

## Stop conditions

Codex must stop implementation and document the blocker if any of these occur:

1. Local repository contains substantial uncommitted work that would be overwritten.
2. The local checkout materially contradicts this handoff and the contradiction cannot be resolved from repository history.
3. Required external services/credentials are unavailable and no local mock substitute is appropriate.
4. Docker/Kind cannot run because of host restrictions outside the repository.
5. A proposed change would require destructive migration of user data.
6. A change would require introducing write/control operations against physical BMCs.
7. A vendor-specific behavior cannot be validated from mock fixtures and would be unsafe to generalize.

When stopping, document:

```text
blocker
evidence
completed work
safe next step
```

Do not fabricate a passing result.

## Do not rewrite without evidence

Avoid replacing an existing subsystem merely because a cleaner abstraction is possible.

A rewrite requires at least one of:

- correctness bug that cannot be fixed locally,
- architectural invariant violation,
- unacceptable concurrency/failure behavior,
- testability blocker,
- clearly lower-risk migration path than patching.

Prefer incremental refactors.

## Preserve compatibility

During Phase 1:

- preserve existing config keys where possible,
- preserve existing metric names where practical,
- preserve read-only behavior,
- preserve current Docker image flow,
- preserve Kind local UX,
- preserve existing OpenSearch indexes unless change is necessary.

If compatibility is intentionally broken, document it explicitly.

## Branch hygiene

Work only on the active feature branch.

Before modifying code:

```bash
git status
git branch --show-current
```

Do not reset, checkout away, clean, stash, or delete user changes unless explicitly instructed.

Do not force-push.

## Secrets

Never commit:

- BMC passwords
- session tokens
- API keys
- private certificates
- production IPs if sensitive
- OpenSearch credentials

Mock credentials should be obviously fake and local-only.

## Test honesty

Never mark PASS based on:

- code inspection alone,
- expected behavior,
- commented-out test,
- skipped test without explanation,
- mocked assertion that bypasses the behavior under test.

Every runtime PASS needs an executed command or observable result.

## Timeouts and flakiness

Use deterministic, generous test thresholds.

Do not "fix" flaky tests by adding arbitrary large sleeps.

Prefer:

```text
poll condition
with timeout
fail with diagnostic
```

## Performance claims

Mock lab performance validates collector mechanics only.

Do not claim real BMC capacity based on local mock throughput.

## Vendor claims

Do not claim compatibility with Dell/HPE/Lenovo/Supermicro unless tested against representative real hardware/firmware or explicitly modeled fixtures.

## Scope control

Phase 1 should not expand into:

- incident correlation,
- predictive analytics,
- firmware compliance policy,
- full EventService implementation,
- BMC power/control actions,
- UI redesign beyond minimum validation needs.
