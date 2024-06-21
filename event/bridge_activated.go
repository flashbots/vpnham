package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeActivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Timestamp       time.Time
}

func (e *BridgeActivated) EventKind() string {
	return "bridge_activated"
}

func (e *BridgeActivated) EventTimestamp() time.Time {
	return e.Timestamp
}
