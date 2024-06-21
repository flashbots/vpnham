package event

import "time"

type PartnerWentDown struct {
	Timestamp time.Time
}

func (e *PartnerWentDown) EventKind() string {
	return "partner_went_down"
}

func (e *PartnerWentDown) EventTimestamp() time.Time {
	return e.Timestamp
}
