package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceDeactivated struct {
	BridgeInterface string
	BridgePeerCIDRs []types.CIDR
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelInterfaceDeactivated) EvtKind() string {
	return "tunnel_interface_deactivated"
}

func (e *TunnelInterfaceDeactivated) EvtBridgeInterface() string {
	return e.BridgeInterface
}

func (e *TunnelInterfaceDeactivated) EvtBridgePeerCIDRs() []types.CIDR {
	return e.BridgePeerCIDRs
}

func (e *TunnelInterfaceDeactivated) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelInterfaceDeactivated) EvtTimestamp() time.Time {
	return e.Timestamp
}
