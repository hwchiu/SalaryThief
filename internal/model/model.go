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

// Inventory keeps physical location separate from mutable hardware identity.
type InventoryComponent struct {
	Type          string      `json:"type"`
	ComponentID   string      `json:"component_id"`
	Location      string      `json:"location"`
	Manufacturer  string      `json:"manufacturer,omitempty"`
	Model         string      `json:"model,omitempty"`
	Serial        string      `json:"serial,omitempty"`
	PartNumber    string      `json:"part_number,omitempty"`
	Firmware      string      `json:"firmware,omitempty"`
	CapacityBytes *uint64     `json:"capacity_bytes,omitempty"`
	Health        HealthState `json:"health"`
}
type InventorySnapshot struct {
	Target     string               `json:"server"`
	Scope      string               `json:"observability_scope"`
	ObservedAt time.Time            `json:"observed_at"`
	Components []InventoryComponent `json:"components"`
}
type InventoryChange struct {
	Target      string    `json:"server"`
	Type        string    `json:"component"`
	ComponentID string    `json:"component_id"`
	Location    string    `json:"location"`
	Change      string    `json:"change"`
	OldSerial   string    `json:"old_serial,omitempty"`
	NewSerial   string    `json:"new_serial,omitempty"`
	ObservedAt  time.Time `json:"observed_at"`
}
