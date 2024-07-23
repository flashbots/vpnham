package reconciler

import (
	"context"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/job"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (r *Reconciler) InterfaceDeactivate(
	ctx context.Context,
	e *event.TunnelInterfaceDeactivated,
	failureSink chan<- error,
) {
	l := logutils.LoggerFromContext(ctx)

	if len(r.cfg.InterfaceDeactivate.Script) == 0 {
		l.Debug("No interface deactivation script configured; skipping...")
		return
	}

	placeholders, err := r.renderPlaceholders(e)
	if err != nil {
		l.Error("Failed to render interface deactivation script",
			zap.Error(err),
		)
		metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, r.name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeSystem),
		))
		return
	}

	r.scheduleJob(job.RunScript(
		"interface_deactivate",
		r.cfg.ScriptsTimeout,
		r.renderScript(&r.cfg.BridgeActivate.Script, placeholders),
	))
}
