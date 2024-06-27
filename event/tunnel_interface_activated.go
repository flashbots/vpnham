package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceActivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelInterfaceActivated) EvtKind() string {
	return "tunnel_interface_activated"
}

func (e *TunnelInterfaceActivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *TunnelInterfaceActivated) EvtBridgePeerCIDR() types.CIDR {
	return e.BridgePeerCIDR
}

func (e *TunnelInterfaceActivated) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelInterfaceActivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
