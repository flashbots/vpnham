package event

import "time"

type TunnelProbeSendSuccess struct {
	Interface string
	Sequence  uint64
	Timestamp time.Time
}

func (e *TunnelProbeSendSuccess) EventKind() string {
	return "tunnel_probe_send_success"
}

func (e *TunnelProbeSendSuccess) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *TunnelProbeSendSuccess) TunnelInterface() string {
	return e.Interface
}
