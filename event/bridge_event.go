package event

import "github.com/flashbots/vpnham/types"

type BridgeEvent interface {
	Event
	EvtBridgeInterface() string
	EvtBridgePeerCIDRs() []types.CIDR
}
