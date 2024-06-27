package event

import (
	"time"
)

type Event interface {
	EvtKind() string
	EvtTimestamp() time.Time
}
