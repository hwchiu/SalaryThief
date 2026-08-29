# Internal Domain Model Contract

This document defines the conceptual model. Codex should adapt it to the existing code instead of blindly reproducing these exact type names.

## Target identity

```go
type TargetID string

type Target struct {
    ID                 TargetID
    Endpoint           string
    ObservabilityScope string
    Vendor             string
    Labels             map[string]string
    // CredentialRef should reference external secret material.
}
```

Requirements:

- target ID must be stable
- target ID must not depend on mutable serial/firmware
- endpoint and credentials are operational attributes, not hardware identity

## Error classification

```go
type ErrorClass string

const (
    ErrorNone    ErrorClass = ""
    ErrorTimeout ErrorClass = "timeout"
    ErrorAuth    ErrorClass = "auth"
    ErrorHTTP    ErrorClass = "http"
    ErrorTLS     ErrorClass = "tls"
    ErrorNetwork ErrorClass = "network"
    ErrorSchema  ErrorClass = "schema"
    ErrorPartial ErrorClass = "partial_resource"
    ErrorUnknown ErrorClass = "unknown"
)
```

## Normalized state

```go
type HealthState int

const (
    HealthUnknown HealthState = iota
    HealthOK
    HealthWarning
    HealthCritical
    HealthError
)
```

`HealthError` means collection/parsing failure, not hardware criticality.

## Resource family

Use bounded families:

```text
system
chassis
manager
thermal
power
processor
memory
storage
network
firmware
logs
boot
```

## Per-resource status

Conceptually:

```go
type ResourceStatus struct {
    State       HealthState
    LastAttempt time.Time
    LastSuccess time.Time
    ErrorClass  ErrorClass
    Error       string // log/debug only; must be sanitized
}
```

Do not expose raw `Error` strings as Prometheus labels.

## Target state

Conceptually:

```go
type TargetState struct {
    TargetID      TargetID
    Connectivity  bool
    LastAttempt   time.Time
    LastSuccess   time.Time
    LastError     ErrorClass

    Resources map[ResourceFamily]ResourceStatus

    Snapshot NormalizedSnapshot
}
```

Concurrency requirements:

- safe concurrent worker updates
- safe concurrent Prometheus reads
- avoid global lock around network calls
- update state atomically enough that readers do not observe impossible combinations

## Normalized snapshot

The normalized snapshot is the boundary between vendor-specific Redfish parsing and sinks.

Conceptually:

```go
type NormalizedSnapshot struct {
    System     SystemInfo
    Components []Component
    Telemetry  []MetricSample
    Inventory  Inventory
    Events     []Event
}
```

The exact shape should reflect the existing codebase.

## Component identity

```go
type Component struct {
    ComponentID string
    Type        string
    Location    string
    Health      HealthState
}
```

Identity fields belong in inventory:

```go
type HardwareIdentity struct {
    Manufacturer string
    Model        string
    Serial       string
    PartNumber   string
    Firmware     string
}
```

Rule:

```text
location != identity
```

Example:

```text
location = Bay 03
serial   = AAA
```

After replacement:

```text
location = Bay 03
serial   = BBB
```

## Inventory model

Conceptual component inventory record:

```go
type InventoryComponent struct {
    ComponentID string
    Type        string
    Location    string
    Identity    HardwareIdentity
    CapacityBytes *uint64
    Health      HealthState
    ObservedAt  time.Time
}
```

Use pointers/optional values for fields that may be absent. Missing numeric data must not silently become zero.

## Event model

Conceptually:

```go
type Event struct {
    Timestamp          time.Time
    TargetID           TargetID
    Severity           string
    Category           string
    EventType          string
    Component          string
    ComponentID        string
    Location           string
    EventID            string
    Fingerprint        string
    Message            string
    RedfishResource    string
}
```

Raw event payload may be stored separately with strict sanitization/retention.

## Collection result

Workers should return a structured result rather than directly mutating every sink.

Conceptually:

```go
type CollectionResult struct {
    TargetID      TargetID
    StartedAt     time.Time
    FinishedAt    time.Time
    Connectivity  bool
    ErrorClass    ErrorClass
    Snapshot      NormalizedSnapshot
    Resources     map[ResourceFamily]ResourceStatus
}
```

Preferred flow:

```text
worker
  ↓
CollectionResult
  ├── state/cache update
  ├── metrics update
  └── persistence queue
```

This simplifies testing and failure isolation.

## State transition rules

### Success after failure

```text
down/stale
  ↓ successful collection
up/current
```

Reset relevant error class and freshness.

### Failure after success

```text
up/current
  ↓ failure
down or resource-error
```

Retain last successful data where useful, but mark stale/error explicitly.

### Partial failure

Only affected resource family changes to ERROR/UNKNOWN.

### Restart

On process restart, in-memory freshness starts empty unless durable state is intentionally added later.

Do not fake historical freshness after restart.
