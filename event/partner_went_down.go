package event

import "time"

type PartnerWentDown struct {
	Timestamp time.Time
}

func (e *PartnerWentDown) EvtKind() string {
	return "partner_went_down"
}

func (e *PartnerWentDown) EvtTimestamp() time.Time {
	return e.Timestamp
}
