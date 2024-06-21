package event

import "time"

type TunnelInterfaceWentDown struct {
	Interface string
	Timestamp time.Time
}

func (e *TunnelInterfaceWentDown) EventKind() string {
	return "tunnel_interface_went_down"
}

func (e *TunnelInterfaceWentDown) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelInterfaceWentDown) TunnelInterface() string {
	return e.Interface
}
