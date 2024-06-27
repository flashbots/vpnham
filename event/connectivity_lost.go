package event

import "time"

type ConnectivityLost struct {
	Timestamp time.Time
}

func (e *ConnectivityLost) EvtKind() string {
	return "connectivity_lost"
}

func (e *ConnectivityLost) EvtTimestamp() time.Time {
	return e.Timestamp
}
