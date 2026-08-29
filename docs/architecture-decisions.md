# Architecture Decisions

## ADR-001 — Decouple Prometheus scrape from BMC polling

**Status:** Accepted

Prometheus scrapes cached collector state. Redfish I/O occurs only in background collection.

Consequences:

- `/metrics` availability is independent of BMC latency
- freshness must be modeled explicitly
- collector owns scheduling and concurrency
- transient BMC failures do not block scrape requests

## ADR-002 — Separate telemetry from inventory/events

**Status:** Accepted

Prometheus stores telemetry and bounded-cardinality health data.

OpenSearch/Elasticsearch stores hardware inventory and event documents.

Consequences:

- do not index full inventory every telemetry cycle
- do not place serial/part/firmware fields in ordinary metrics labels
- OpenSearch outage must be isolated from Prometheus scrape path

## ADR-003 — Separate physical location from hardware identity

**Status:** Accepted

A bay/slot/socket is a physical location. A serial-numbered device is hardware identity.

Consequences:

- replacements can be detected
- historical RMA evidence is meaningful
- `Bay 03` is not itself a unique disk identity

## ADR-004 — Mock lab validates architecture, not vendor certification

**Status:** Accepted

Kind + Redfish mocks are the primary local test harness.

Consequences:

- protocol/data-flow/failure behavior can be automated locally
- OEM compatibility claims require later real hardware/firmware testing
