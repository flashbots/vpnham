package event

import "time"

type PartnerActivated struct {
	Timestamp time.Time
}

func (e *PartnerActivated) EvtKind() string {
	return "partner_activated"
}

func (e *PartnerActivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
