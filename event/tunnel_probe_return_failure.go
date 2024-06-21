package event

import "time"

type TunnelProbeReturnFailure struct {
	Interface string
	Sequence  uint64
	Timestamp time.Time
}

func (e *TunnelProbeReturnFailure) EventKind() string {
	return "tunnel_probe_return_failure"
}

func (e *TunnelProbeReturnFailure) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelProbeReturnFailure) TunnelInterface() string {
	return e.Interface
}
