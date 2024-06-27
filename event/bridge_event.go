package event

import "github.com/flashbots/vpnham/types"

type BridgeEvent interface {
	Event
	EvtBridgeInterface() string
	EvtBridgePeerCIDR() types.CIDR
}
