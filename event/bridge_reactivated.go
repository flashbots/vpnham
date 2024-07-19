package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeReactivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Iteration       int
	Timestamp       time.Time
}

func (e *BridgeReactivated) EvtKind() string {
	return "bridge_reactivated"
}

func (e *BridgeReactivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeReactivated) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}

func (e *BridgeReactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
