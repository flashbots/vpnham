package event

import "time"

type PartnerChangedName struct {
	NewName   string
	OldName   string
	Timestamp time.Time
}

func (e *PartnerChangedName) EvtKind() string {
	return "partner_changed_name"
}

func (e *PartnerChangedName) EvtTimestamp() time.Time {
	return e.Timestamp
}
