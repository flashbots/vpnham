package metrics

import (
	otelapi "go.opentelemetry.io/otel/metric"
)

// Bridge

var (
	// BridgeActive indicates whether the bridge is currently active
	BridgeActive otelapi.Int64Observable

	// BridgeUp indicates whether the bridge is currently up
	BridgeUp otelapi.Int64Observable

	// TunnelInterfaceActive indicates whether the tunnel interface is active
	TunnelInterfaceActive otelapi.Int64Observable

	// TunnelInterfaceUp indicates whether the tunnel interface is up
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
