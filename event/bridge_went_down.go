package event

import "time"

type BridgeWentDown struct {
	Timestamp time.Time
}

func (e *BridgeWentDown) EventKind() string {
	return "bridge_went_down"
}

func (e *BridgeWentDown) EventTimestamp() time.Time {
	return e.Timestamp
}
