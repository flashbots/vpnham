package bridge

import (
	"context"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/monitor"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
)

func (s *Server) eventTunnelProbeSendSuccess(ctx context.Context, e *event.TunnelProbeSendSuccess, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.ProbeSequence, monitor.Pending)
	})

	metrics.ProbesSent.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
	))
}

func (s *Server) eventTunnelProbeSendFailure(ctx context.Context, e *event.TunnelProbeSendFailure, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.ProbeSequence, monitor.Down)
	})

	metrics.ProbesFailed.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
	))
}

func (s *Server) eventTunnelProbeReturnSuccess(ctx context.Context, e *event.TunnelProbeReturnSuccess, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.ProbeSequence, monitor.Up)
	})

	metrics.ProbesReturned.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
	))

	metrics.ProbesLatencyForward.Record(ctx, float64(e.LatencyForward.Milliseconds()), otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
		attribute.String(metrics.LabelProbeDst, e.Location),
		attribute.String(metrics.LabelProbeSrc, s.cfg.ProbeLocation.String()),
	))

	metrics.ProbesLatencyReturn.Record(ctx, float64(e.LatencyReturn.Milliseconds()), otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
		attribute.String(metrics.LabelProbeDst, s.cfg.ProbeLocation.String()),
		attribute.String(metrics.LabelProbeSrc, e.Location),
	))
}

func (s *Server) eventTunnelProbeReturnFailure(ctx context.Context, e *event.TunnelProbeReturnFailure, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.ProbeSequence, monitor.Down)
	})

	metrics.ProbesFailed.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String(metrics.LabelBridge, s.cfg.Name),
		attribute.String(metrics.LabelTunnel, e.TunnelInterface),
	))
}
