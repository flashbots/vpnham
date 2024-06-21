package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type PartnerPollFailure struct {
	Sequence  uint64
	Timestamp time.Time
}

func (e *PartnerPollFailure) EventKind() string {
	return "partner_poll_failure"
}

func (e *PartnerPollFailure) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *PartnerPollFailure) PartnerStatus() *types.BridgeStatus {
	return nil
}
