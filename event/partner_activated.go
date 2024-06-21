package event

import "time"

type PartnerActivated struct {
	Timestamp time.Time
}

func (e *PartnerActivated) EventKind() string {
	return "partner_activated"
}

func (e *PartnerActivated) EventTimestamp() time.Time {
	return e.Timestamp
}
