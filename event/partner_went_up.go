package event

import "time"

type PartnerWentUp struct {
	Timestamp time.Time
}

func (e *PartnerWentUp) EventKind() string {
	return "partner_went_down"
}

func (e *PartnerWentUp) EventTimestamp() time.Time {
	return e.Timestamp
}
