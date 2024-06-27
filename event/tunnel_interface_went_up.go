package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceWentUp struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelInterfaceWentUp) EvtKind() string {
	return "tunnel_interface_went_up"
}

func (e *TunnelInterfaceWentUp) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *TunnelInterfaceWentUp) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}
func (e *TunnelInterfaceWentUp) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelInterfaceWentUp) EvtTimestamp() time.Time {
	return e.Timestamp
}
