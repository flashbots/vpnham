package event

import "time"

type TunnelProbeSendSuccess struct {
	ProbeSequence   uint64
	TunnelInterface string
	Timestamp       time.Time
}

func (e *TunnelProbeSendSuccess) EvtKind() string {
	return "tunnel_probe_send_success"
}

func (e *TunnelProbeSendSuccess) EvtTunnelInterface() string {
	return e.TunnelInterface
}

func (e *TunnelProbeSendSuccess) EvtTimestamp() time.Time {
	return e.Timestamp
}
