package event

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type TunnelInterfaceActivated struct {
	BridgeInterface string
	BridgePeerCIDR  types.CIDR
	Interface       string
	Timestamp       time.Time
}

func (e *TunnelInterfaceActivated) EventKind() string {
	return "tunnel_interface_activated"
}

func (e *TunnelInterfaceActivated) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelInterfaceActivated) TunnelInterface() string {
	return e.Interface
}
