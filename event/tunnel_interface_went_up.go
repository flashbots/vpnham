package event

import "time"

type TunnelInterfaceWentUp struct {
	Interface string
	Timestamp time.Time
}

func (e *TunnelInterfaceWentUp) EventKind() string {
	return "tunnel_interface_went_up"
}

func (e *TunnelInterfaceWentUp) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelInterfaceWentUp) TunnelInterface() string {
	return e.Interface
}
