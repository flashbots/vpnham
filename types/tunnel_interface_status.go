package types

import "time"

type TunnelInterfaceStatus struct {
	// Active indicates whether the tunnel is currently in active state.
	Active bool `json:"active"`

	// ActiveSince is the timestamp of most recent update to the Active state
	// (regardless whether true or false).
	ActiveSince time.Time `json:"active_since"`

	// Up indicates whether the tunnel is currently in up (online) state.
	Up bool `json:"up"`

	// UpSince is the timestamp of most recent update to the Up state
	// (regardless whether true or false).
	UpSince time.Time `json:"up_since"`
}
