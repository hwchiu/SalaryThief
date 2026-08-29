package model

import "time"

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

type HealthState int

const (
	HealthUnknown HealthState = iota
	HealthOK
	HealthWarning
	HealthCritical
	HealthError
)

type ResourceStatus struct {
	State       HealthState
	LastAttempt time.Time
	LastSuccess time.Time
	ErrorClass  ErrorClass
}

// Snapshot is the normalized, cacheable result of one target collection.
// It deliberately excludes credentials and raw Redfish payloads.
type Snapshot struct {
	Target      string
	Scope       string
	Labels      map[string]string
	Up          bool
	LastAttempt time.Time
	LastSuccess time.Time
	ErrorClass  ErrorClass
	Duration    time.Duration
	Resources   map[string]ResourceStatus
}
