package event

import "time"

type ConnectivityRestored struct {
	Timestamp time.Time
}

func (e *ConnectivityRestored) EvtKind() string {
	return "connectivity_restored"
}

func (e *ConnectivityRestored) EvtTimestamp() time.Time {
	return e.Timestamp
}
