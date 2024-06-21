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
		m.RegisterStatus(e.Sequence, monitor.Pending)
	})

	metrics.ProbesSent.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
	))
}

func (s *Server) eventTunnelProbeSendFailure(ctx context.Context, e *event.TunnelProbeSendFailure, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.Sequence, monitor.Down)
	})

	metrics.ProbesFailed.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
	))
}

func (s *Server) eventTunnelProbeReturnSuccess(ctx context.Context, e *event.TunnelProbeReturnSuccess, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.Sequence, monitor.Up)
	})

	metrics.ProbesReturned.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
	))

	metrics.ProbesLatencyForward.Record(ctx, float64(e.LatencyForward.Milliseconds()), otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
		attribute.String("location_from", s.cfg.ProbeLocation.String()),
		attribute.String("location_to", e.Location),
	))

	metrics.ProbesLatencyReturn.Record(ctx, float64(e.LatencyReturn.Milliseconds()), otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
		attribute.String("location_from", e.Location),
		attribute.String("location_to", s.cfg.ProbeLocation.String()),
	))
}

func (s *Server) eventTunnelProbeReturnFailure(ctx context.Context, e *event.TunnelProbeReturnFailure, _ chan<- error) {
	s.detectTunnelUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.Sequence, monitor.Down)
	})

	metrics.ProbesFailed.Add(ctx, 1, otelapi.WithAttributes(
		attribute.String("bridge_name", s.cfg.Name),
		attribute.String("interface_name", e.Interface),
	))
}
