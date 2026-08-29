# SalaryThief v2 Architecture

## Logical components

```text
                          ┌──────────────────────┐
                          │    Server Registry   │
                          └──────────┬───────────┘
                                     │
                          ┌──────────▼───────────┐
                          │      Scheduler       │
                          └──────┬──────┬────────┘
                                 │      │
                 ┌───────────────┘      └──────────────┐
                 │                                      │
        ┌────────▼────────┐                    ┌────────▼────────┐
        │ Telemetry Queue │                    │ Event/Inv Queues │
        └────────┬────────┘                    └────────┬────────┘
                 │                                      │
        ┌────────▼──────────────────────────────────────▼────────┐
        │              Bounded Worker Pools                      │
        └───────────────────────┬─────────────────────────────────┘
                                │
                       ┌────────▼────────┐
                       │ Redfish Clients │
                       │ + vendor/OEM    │
                       │ adapters        │
                       └────────┬────────┘
                                │
                       ┌────────▼────────┐
                       │      BMCs       │
                       └────────┬────────┘
                                │
                       ┌────────▼────────┐
                       │   Normalizer    │
                       └───┬─────────┬───┘
                           │         │
                ┌──────────▼───┐ ┌──▼────────────────┐
                │ Metrics Cache│ │ OpenSearch Writers│
                └──────┬───────┘ │ events/inventory │
                       │         └───────────────────┘
                ┌──────▼───────┐
                │   /metrics   │
                └──────┬───────┘
                       │
                ┌──────▼───────┐
                │  Prometheus  │
                └──────────────┘
```

## Key separation

The scrape path is cache-only. Redfish network I/O is always performed asynchronously by background collection.

## Internal state

A practical state record per target should be able to represent:

- last attempt timestamp
- last successful collection timestamp
- last successful data per resource family
- current connectivity state
- per-resource status (OK/WARNING/CRITICAL/UNKNOWN/ERROR)
- data age/staleness
- last error class
- current retry/backoff state

Do not clear all previously good component state merely because one child resource failed.

## Vendor boundary

Keep generic collection/normalization interfaces independent of Dell/HPE/Lenovo/Supermicro OEM payloads.

Conceptually:

```text
raw Redfish / OEM payload
          ↓
standard resource parser + vendor adapter
          ↓
normalized internal model
          ↓
metrics/events/inventory
```

## Persistence boundaries

Prometheus is not an inventory database. OpenSearch is not the synchronous metrics serving path.

An OpenSearch outage must be observable and recoverable without breaking cached Prometheus metrics.

## Target scale

Plan and benchmark for at least:

```text
200
1,000
2,000
5,000
10,000 BMCs
```

The local mock environment can generate large logical fleets to exercise scheduler/queue/cache behavior, but mock throughput must never be interpreted as real BMC capacity.
