package event

import "time"

type TunnelProbeSendFailure struct {
	ProbeSequence   uint64
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelProbeSendFailure) EvtKind() string {
	return "tunnel_probe_send_failure"
}

func (e *TunnelProbeSendFailure) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelProbeSendFailure) EvtTimestamp() time.Time {
	return e.Timestamp
}
