package bridge

import (
	"context"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
)

func (s *Server) ObserveMetrics(ctx context.Context, observer otelapi.Observer) error {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Observing metrics")

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	s.mxPartnerStatus.Lock()
	defer s.mxPartnerStatus.Unlock()

	{ // bridge_active
		var val int64 = 0
		if s.status.Active {
			val++
		}
		if s.partnerStatus != nil && s.partnerStatus.Active {
			val++
		}
		observer.ObserveInt64(metrics.BridgeActive, val, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, s.cfg.Name),
		))
	}

	{ // bridge_active
		var val int64 = 0
		if s.status.Up {
			val++
		}
		if s.partnerStatus != nil && s.partnerStatus.Up {
			val++
		}
		observer.ObserveInt64(metrics.BridgeUp, val, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, s.cfg.Name),
		))
	}

	for ifsName, ifs := range s.status.Interfaces {
		{ // tunnel_interface_active
			var val int64 = 0
			if ifs.Active {
				val = 1
			}
			observer.ObserveInt64(metrics.TunnelInterfaceActive, val, otelapi.WithAttributes(
				attribute.String(metrics.LabelBridge, s.cfg.Name),
				attribute.String(metrics.LabelTunnel, ifsName),
			))
		}

		{ // tunnel_interface_up
			var val int64 = 0
			if ifs.Up {
				val = 1
			}
			observer.ObserveInt64(metrics.TunnelInterfaceUp, val, otelapi.WithAttributes(
				attribute.String(metrics.LabelBridge, s.cfg.Name),
				attribute.String(metrics.LabelTunnel, ifsName),
			))
		}
	}

	return nil
}
