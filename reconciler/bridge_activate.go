package reconciler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/job"
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
	r.bridgeActivateUpdateGCP(ctx, e)
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
	aws := r.cfg.BridgeActivate.AWS

	for _, vpc := range aws.Vpcs {
		r.scheduleJob(&job.UpdateAWSRouteTables{
			JobName: "aws_update_route_tables",
			Timeout: aws.Timeout,

			DestinationCidrBlocks: e.EvtBridgePeerCIDRs(),
			NetworkInterfaceID:    vpc.NetworkInterfaceID,
			RouteTables:           vpc.RouteTables,
		})
	}
}

func (r *Reconciler) bridgeActivateUpdateGCP(
	ctx context.Context,
	e event.BridgeEvent,
) {
	l := logutils.LoggerFromContext(ctx)

	if r.cfg.BridgeActivate.GCP == nil {
		l.Debug("No bridge activation GCP configuration provided; skipping...")
		return
	}
	gcp := r.cfg.BridgeActivate.GCP

	for id, vpc := range gcp.Vpcs {
		parts := strings.Split(id, "/")
		name := gcp.RouteIDPrefix + "-" + parts[len(parts)-1]

		description := "Created by vpnham on " + time.Now().UTC().Format(time.RFC3339)

		r.scheduleJob(&job.UpdateGCPRoute{
			JobName: "gcp_update_route",
			Timeout: gcp.Timeout,

			Name:        name,
			Description: description,

			DestRanges:      e.EvtBridgePeerCIDRs(),
			Network:         vpc.ID,
			NextHopInstance: gcp.InstanceName,
			Priority:        gcp.RoutePriority,
			Tags:            gcp.RouteTags,
		})
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

	r.scheduleJob(&job.RunScript{
		JobName: "bridge_activate",
		Timeout: r.cfg.ScriptsTimeout,

		Script: r.renderScript(&r.cfg.BridgeActivate.Script, placeholders),
	})
}
