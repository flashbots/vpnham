package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeWentDown struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Timestamp       time.Time
}

func (e *BridgeWentDown) EvtKind() string {
	return "bridge_went_down"
}

func (e *BridgeWentDown) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeWentDown) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}

func (e *BridgeWentDown) EvtTimestamp() time.Time {
	return e.Timestamp
}
