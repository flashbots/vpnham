package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type BridgeWentUp struct {
	BridgeInterface string
	BridgePerCIDR   types.CIDR
	Timestamp       time.Time
}

func (e *BridgeWentUp) EvtKind() string {
	return "bridge_went_up"
}

func (e *BridgeWentUp) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *BridgeWentUp) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePerCIDR
}

func (e *BridgeWentUp) EvtTimestamp() time.Time {
	return e.Timestamp
}
