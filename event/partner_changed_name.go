package event

import "time"

type PartnerChangedName struct {
	NewName   string
	OldName   string
	Timestamp time.Time
}

func (e *PartnerChangedName) EventKind() string {
	return "partner_changed_name"
}

func (e *PartnerChangedName) EventTimestamp() time.Time {
	return e.Timestamp
}
