package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeActivated struct {
	BridgeInterface string
	BridgePeerCIDRs []types.CIDR
	Timestamp       time.Time
}

func (e *BridgeActivated) EvtKind() string {
	return "bridge_activated"
}

func (e *BridgeActivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeActivated) EvtBridgePeerCIDRs() []types.CIDR {
	return e.BridgePeerCIDRs
}

func (e *BridgeActivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
