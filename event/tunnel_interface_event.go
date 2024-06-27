package event

type TunnelInterfaceEvent interface {
	Event
	EvtTunnelInterface() string
}
