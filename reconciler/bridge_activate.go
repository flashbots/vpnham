package reconciler

import (
	"context"
	"fmt"

	"github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (r *Reconciler) BridgeActivate(
	ctx context.Context,
	e event.BridgeEvent,
	failureSink chan<- error,
) {
	switch e.(type) {
	case *event.BridgeActivated:
	case *event.BridgeReactivated:
		// pass
	default:
		failureSink <- fmt.Errorf("unexpected event is trying to (re-)activate the bridge: %s",
			e.EvtKind(),
		)
	}

	r.bridgeActivateUpdateAWS(ctx, e)
	r.bridgeActivateRunScript(ctx, e)
}

func (r *Reconciler) bridgeActivateUpdateAWS(
	ctx context.Context,
	e event.BridgeEvent,
) {
	l := logutils.LoggerFromContext(ctx)

	if r.cfg.BridgeActivate.AWS == nil {
		l.Debug("No bridge activation AWS configuration provided; skipping...")
		return
	}

	cli, err := aws.NewClient(ctx)
	if err != nil {
		l.Error("Failed to update AWS",
			zap.Error(err),
		)
		metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, r.name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeAWS),
		))
		return
	}

	cidr := e.EvtBridgePeerCIDR().String()
	networkInterfaceID := r.cfg.BridgeActivate.AWS.NetworkInterfaceID

	for _, rt := range r.cfg.BridgeActivate.AWS.RouteTables {
		err := cli.UpdateRouteTable(ctx, rt, cidr, networkInterfaceID)
		if err != nil {
			l.Error("Failed to update AWS route table",
				zap.Error(err),
				zap.String("network_interface_id", networkInterfaceID),
				zap.String("route_table", rt),
			)
			metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
				attribute.String(metrics.LabelBridge, r.name),
				attribute.String(metrics.LabelErrorScope, metrics.ScopeAWS),
			))
		}
	}
}

func (r *Reconciler) bridgeActivateRunScript(
	ctx context.Context,
	e event.BridgeEvent,
) {
	l := logutils.LoggerFromContext(ctx)

	if len(r.cfg.BridgeActivate.Script) == 0 {
		l.Debug("No bridge activation script configured; skipping...")
		return
	}

	placeholders, err := r.renderPlaceholders(e)
	if err != nil {
		l.Error("Failed to render interface activation script",
			zap.Error(err),
		)
		metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, r.name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeSystem),
		))
		return
	}

	r.scheduleJob(job{
		name:   "bridge_activate",
		script: r.renderScript(&r.cfg.BridgeActivate.Script, placeholders),
	})
}
