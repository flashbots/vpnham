package event

import "time"

type PartnerDeactivated struct {
	Timestamp time.Time
}

func (e *PartnerDeactivated) EvtKind() string {
	return "partner_deactivated"
}

func (e *PartnerDeactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
