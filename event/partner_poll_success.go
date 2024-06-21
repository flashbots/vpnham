package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type PartnerPollSuccess struct {
	Status    *types.BridgeStatus
	Sequence  uint64
	Timestamp time.Time
}

func (e *PartnerPollSuccess) EventKind() string {
	return "partner_poll_success"
}

func (e *PartnerPollSuccess) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *PartnerPollSuccess) PartnerStatus() *types.BridgeStatus {
	return e.Status
}
