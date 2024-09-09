package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceReactivated struct {
	BridgeInterface string
	BridgePeerCIDRs []types.CIDR
	Iteration       int
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelInterfaceReactivated) EvtKind() string {
	return "tunnel_interface_reactivated"
}

func (e *TunnelInterfaceReactivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *TunnelInterfaceReactivated) EvtBridgePeerCIDRs() []types.CIDR {
	return e.BridgePeerCIDRs
}

func (e *TunnelInterfaceReactivated) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelInterfaceReactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
