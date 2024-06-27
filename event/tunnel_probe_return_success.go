package event

import "time"

type TunnelProbeReturnSuccess struct {
	ProbeSequence   uint64
	TunnelInterface string
	Timestamp       time.Time

	LatencyForward time.Duration
	LatencyReturn  time.Duration
	Location       string
}

func (e *TunnelProbeReturnSuccess) EvtKind() string {
	return "tunnel_probe_return_success"
}

func (e *TunnelProbeReturnSuccess) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelProbeReturnSuccess) EvtTimestamp() time.Time {
	return e.Timestamp
}
