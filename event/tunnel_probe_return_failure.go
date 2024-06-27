package event

import "time"

type TunnelProbeReturnFailure struct {
	ProbeSequence   uint64
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelProbeReturnFailure) EvtKind() string {
	return "tunnel_probe_return_failure"
}

func (e *TunnelProbeReturnFailure) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelProbeReturnFailure) EvtTimestamp() time.Time {
	return e.Timestamp
}
