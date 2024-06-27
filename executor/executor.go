package executor

import (
	"context"
	"sync"
	"time"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Executor struct {
	bridgeName string
	bridgeUUID uuid.UUID

	bridgeActivate      types.Script
	interfaceActivate   types.Script
	interfaceDeactivate types.Script

	timeout time.Duration

	queue   []job
	mxQueue sync.Mutex

	next chan job
	stop chan struct{}
}

func New(cfg *config.Bridge) (*Executor, error) {
	return &Executor{
		bridgeName: cfg.Name,
		bridgeUUID: cfg.UUID,

		bridgeActivate:      cfg.Scripts.BridgeActivate,
		interfaceActivate:   cfg.Scripts.InterfaceActivate,
		interfaceDeactivate: cfg.Scripts.InterfaceDeactivate,

		timeout: cfg.ScriptsTimeout,

		queue: make([]job, 0, 1),

		next: make(chan job),
		stop: make(chan struct{}, 1),
	}, nil
}

func (ex *Executor) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("VPN HA-monitor script-executor is going up...",
		zap.String("bridge_name", ex.bridgeName),
	)

	go func() {
		ex.loop(ctx)

		l.Info("VPN HA-monitor script-executor is down",
			zap.String("bridge_name", ex.bridgeName),
		)
	}()
}

func (ex *Executor) Stop(ctx context.Context) {
	ex.stop <- struct{}{}
}

func (ex *Executor) ExecuteBridgeActivate(ctx context.Context, e *event.BridgeActivated) {
	l := logutils.LoggerFromContext(ctx)

	if len(ex.bridgeActivate) == 0 {
		l.Debug("No bridge activation script configured; skipping...")
		return
	}

	script, err := func() (*types.Script, error) {
		placeholders, err := ex.renderPlaceholders(e)
		if err != nil {
			return nil, err
		}

		return ex.render(&ex.bridgeActivate, placeholders), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
		)
		return
	}

	ex.schedule(job{
		name:   "bridge_activate",
		script: script,
	})
}

func (ex *Executor) ExecuteInterfaceActivate(ctx context.Context, e *event.TunnelInterfaceActivated) {
	l := logutils.LoggerFromContext(ctx)

	if len(ex.interfaceActivate) == 0 {
		l.Debug("No interface activation script configured; skipping...")
		return
	}

	script, err := func() (*types.Script, error) {
		placeholders, err := ex.renderPlaceholders(e)
		if err != nil {
			return nil, err
		}

		return ex.render(&ex.interfaceActivate, placeholders), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
		)
		return
	}

	ex.schedule(job{
		name:   "interface_activate",
		script: script,
	})
}

func (ex *Executor) ExecuteInterfaceDeactivate(ctx context.Context, e *event.TunnelInterfaceDeactivated) {
	l := logutils.LoggerFromContext(ctx)

	if len(ex.interfaceDeactivate) == 0 {
		l.Debug("No interface deactivation script configured; skipping...")
		return
	}

	script, err := func() (*types.Script, error) {
		placeholders, err := ex.renderPlaceholders(e)
		if err != nil {
			return nil, err
		}

		return ex.render(&ex.interfaceDeactivate, placeholders), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
		)
		return
	}

	ex.schedule(job{
		name:   "interface_deactivate",
		script: script,
	})
}

func (ex *Executor) renderPlaceholders(e event.Event) (map[string]string, error) {
	placeholders := map[string]string{}
	var err error

	if e, ok := e.(event.BridgeEvent); ok {
		placeholders[placeholderBridgeInterface] = e.EvtBridgeInterface()
		placeholders[placeholderBridgePeerCIDR] = string(e.EvtBridgePeerCIDR())

		ipv4 := e.EvtBridgePeerCIDR().IsIPv4()
		if ipv4 {
			placeholders[placeholderProto] = "4"
		} else {
			placeholders[placeholderProto] = "6"
		}

		placeholders[placeholderBridgeInterfaceIP], err = utils.GetInterfaceIP(e.EvtBridgeInterface(), ipv4)
		if err != nil {
			return nil, err
		}

		if e, ok := e.(event.TunnelInterfaceEvent); ok {
			placeholders[placeholderTunnelInterface] = e.EvtTunnelInterface()

			placeholders[placeholderTunnelInterfaceIP], err = utils.GetInterfaceIP(e.EvtTunnelInterface(), ipv4)
			if err != nil {
				return nil, err
			}
		}
	}

	return placeholders, nil
}
