package event

import "time"

type BridgeWentUp struct {
	Timestamp time.Time
}

func (e *BridgeWentUp) EventKind() string {
	return "bridge_went_up"
}

func (e *BridgeWentUp) EventTimestamp() time.Time {
	return e.Timestamp
}
