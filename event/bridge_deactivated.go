package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeDeactivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Timestamp       time.Time
}

func (e *BridgeDeactivated) EvtKind() string {
	return "bridge_deactivated"
}

func (e *BridgeDeactivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeDeactivated) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}

func (e *BridgeDeactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
