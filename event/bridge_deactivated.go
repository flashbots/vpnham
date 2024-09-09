package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeDeactivated struct {
	BridgeInterface string
	BridgePeerCIDRs []types.CIDR
	Timestamp       time.Time
}

func (e *BridgeDeactivated) EvtKind() string {
	return "bridge_deactivated"
}

func (e *BridgeDeactivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeDeactivated) EvtBridgePeerCIDRs() []types.CIDR {
	return e.BridgePeerCIDRs
}

func (e *BridgeDeactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
