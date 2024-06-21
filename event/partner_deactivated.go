package event

import "time"

type PartnerDeactivated struct {
	Timestamp time.Time
}

func (e *PartnerDeactivated) EventKind() string {
	return "partner_deactivated"
}

func (e *PartnerDeactivated) EventTimestamp() time.Time {
	return e.Timestamp
}
