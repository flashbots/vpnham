package event

import "time"

type TunnelProbeSendFailure struct {
	Interface string
	Sequence  uint64
	Timestamp time.Time
}

func (e *TunnelProbeSendFailure) EventKind() string {
	return "tunnel_probe_send_failure"
}

func (e *TunnelProbeSendFailure) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelProbeSendFailure) TunnelInterface() string {
	return e.Interface
}
