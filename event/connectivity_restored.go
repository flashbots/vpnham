package event

import "time"

type ConnectivityRestored struct {
	Timestamp time.Time
}

func (e *ConnectivityRestored) EventKind() string {
	return "connectivity_restored"
}

func (e *ConnectivityRestored) EventTimestamp() time.Time {
	return e.Timestamp
}
