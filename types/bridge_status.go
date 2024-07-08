package types

import "time"

type BridgeStatus struct {
	// Name is the name of the bridge.  Must match the name of the partner's
	// bridge.
	Name string `json:"name"`

	// Role is the configured role of the bridge.
	Role Role `json:"role"`

	// Active indicates whether the bridge is currently in active state.
	Active bool `json:"active"`

	// ActiveSince is the timestamp of most recent update to the Active state
	// (regardless whether true or false).
	ActiveSince time.Time `json:"active_since"`

	// Up indicates whether the bridge is currently in up (online) state.
	Up bool `json:"up"`

	// UpSince is the timestamp of most recent update to the Up state
	// (regardless whether true or false).
	UpSince time.Time `json:"up_since"`

	// Interfaces is the dictionary with bridge interface statuses.
	Interfaces map[string]*TunnelInterfaceStatus `json:"interfaces"`
}
