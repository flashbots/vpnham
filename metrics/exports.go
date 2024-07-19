package metrics

import (
	otelapi "go.opentelemetry.io/otel/metric"
)

// Bridge

var (
	// BridgeActive is a number of active bridges at a given moment
	BridgeActive otelapi.Int64Observable

	// BridgeUp is a number of online bridges at a given moment
	BridgeUp otelapi.Int64Observable

	// TunnelInterfaceActive is a number of active tunnel interfaces at a given
	// moment
	TunnelInterfaceActive otelapi.Int64Observable

	// TunnelInterfaceUp is a number of online tunnel interfaces at a given
	// moment
	TunnelInterfaceUp otelapi.Int64Observable
)

// Probes

var (
	// ProbesSent is a counter for the sent probes
	ProbesSent otelapi.Int64Counter

	// ProbesReturned is a counter for the returned probes
	ProbesReturned otelapi.Int64Counter

	// ProbesFailed is a counter for the failed probes
	ProbesFailed otelapi.Int64Counter

	// ProbesLatencyForward is the latency of probes on the way there
	ProbesLatencyForward otelapi.Float64Histogram

	// ProbesLatencyReturn is the latency of probes on the way back
	ProbesLatencyReturn otelapi.Float64Histogram
)

// Errors

var (
	// Errors is a counter for the errors
	Errors otelapi.Int64Counter
)
