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

func (e *PartnerPollSuccess) EvtKind() string {
	return "partner_poll_success"
}

func (e *PartnerPollSuccess) PartnerStatus() *types.BridgeStatus {
	return e.Status
}

func (e *PartnerPollSuccess) EvtTimestamp() time.Time {
	return e.Timestamp
}
