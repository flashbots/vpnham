package reconciler

import (
	"context"
	"fmt"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/job"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (r *Reconciler) InterfaceActivate(
	ctx context.Context,
	e event.TunnelInterfaceEvent,
	failureSink chan<- error,
) {
	switch e.(type) {
	case *event.TunnelInterfaceActivated:
	case *event.TunnelInterfaceReactivated:
		// pass
	default:
		failureSink <- fmt.Errorf("unexpected event is trying to (re-)activate the interface: %s",
			e.EvtKind(),
		)
	}

	r.interfaceActivateRunScript(ctx, e)
}

func (r *Reconciler) interfaceActivateRunScript(
	ctx context.Context,
	e event.TunnelInterfaceEvent,
) {
	l := logutils.LoggerFromContext(ctx)

	if len(r.cfg.InterfaceActivate.Script) == 0 {
		l.Debug("No interface activation script configured; skipping...")
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

	r.scheduleJob(job.RunScript(
		"interface_activate",
		r.cfg.ScriptsTimeout,
		r.renderScript(&r.cfg.InterfaceActivate.Script, placeholders),
	))
}
