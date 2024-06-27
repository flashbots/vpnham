package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceWentDown struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelInterfaceWentDown) EvtKind() string {
	return "tunnel_interface_went_down"
}

func (e *TunnelInterfaceWentDown) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *TunnelInterfaceWentDown) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}

func (e *TunnelInterfaceWentDown) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelInterfaceWentDown) EvtTimestamp() time.Time {
	return e.Timestamp
}
