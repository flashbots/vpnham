package event

import "time"

type TunnelProbeReturnSuccess struct {
	Interface      string
	LatencyForward time.Duration
	LatencyReturn  time.Duration
	Location       string
	Sequence       uint64
	Timestamp      time.Time
}

func (e *TunnelProbeReturnSuccess) EventKind() string {
	return "tunnel_probe_return_success"
}

func (e *TunnelProbeReturnSuccess) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelProbeReturnSuccess) TunnelInterface() string {
	return e.Interface
}
