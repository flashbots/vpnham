package event

import "time"

type PartnerWentUp struct {
	Timestamp time.Time
}

func (e *PartnerWentUp) EvtKind() string {
	return "partner_went_down"
}

func (e *PartnerWentUp) EvtTimestamp() time.Time {
	return e.Timestamp
}
