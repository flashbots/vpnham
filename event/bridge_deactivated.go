package event

import "time"

type BridgeDeactivated struct {
	Timestamp time.Time
}

func (e *BridgeDeactivated) EventKind() string {
	return "bridge_deactivated"
}

func (e *BridgeDeactivated) EventTimestamp() time.Time {
	return e.Timestamp
}
