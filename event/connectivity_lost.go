package event

import "time"

type ConnectivityLost struct {
	Timestamp time.Time
}

func (e *ConnectivityLost) EventKind() string {
	return "connectivity_lost"
}

func (e *ConnectivityLost) EventTimestamp() time.Time {
	return e.Timestamp
}
