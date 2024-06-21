package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceDeactivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Interface       string
	Timestamp       time.Time
}

func (e *TunnelInterfaceDeactivated) EventKind() string {
	return "tunnel_interface_deactivated"
}

func (e *TunnelInterfaceDeactivated) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelInterfaceDeactivated) TunnelInterface() string {
	return e.Interface
}
