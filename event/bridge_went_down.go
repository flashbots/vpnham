package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeWentDown struct {
	BridgeInterface string
	BridgePeerCIDRs []types.CIDR
	Timestamp       time.Time
}

func (e *BridgeWentDown) EvtKind() string {
	return "bridge_went_down"
}

func (e *BridgeWentDown) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeWentDown) EvtBridgePeerCIDRs() []types.CIDR {
	return e.BridgePeerCIDRs
}

func (e *BridgeWentDown) EvtTimestamp() time.Time {
	return e.Timestamp
}
