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

	queue   []*types.Script
	mxQueue sync.Mutex

	next chan *types.Script
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

		queue: make([]*types.Script, 0, 1),

		next: make(chan *types.Script),
		stop: make(chan struct{}, 1),
	}, nil
}

func (ex *Executor) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("VPN HA-monitor script-executor is going up...",
		zap.String("bridge_name", ex.bridgeName),
		zap.String("bridge_uuid", ex.bridgeUUID.String()),
	)

	go func() {
		ex.loop(ctx)

		l.Info("VPN HA-monitor script-executor is down",
			zap.String("bridge_name", ex.bridgeName),
			zap.String("bridge_uuid", ex.bridgeUUID.String()),
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
		bridgeInterfaceIP, err := utils.GetInterfaceIP(e.BridgeInterface, e.BridgePeerCIDR.IsIPv4())
		if err != nil {
			return nil, err
		}
		return ex.render(&ex.bridgeActivate, map[string]string{
			placeholderBridgeInterface:   e.BridgeInterface,
			placeholderBridgeInterfaceIP: bridgeInterfaceIP,
			placeholderBridgePeerCIDR:    string(e.BridgePeerCIDR),
		}), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
			zap.String("bridge_uuid", ex.bridgeUUID.String()),
		)
		return
	}

	ex.schedule(script)
}

func (ex *Executor) ExecuteInterfaceActivate(ctx context.Context, e *event.TunnelInterfaceActivated) {
	l := logutils.LoggerFromContext(ctx)

	if len(ex.interfaceActivate) == 0 {
		l.Debug("No interface activation script configured; skipping...")
		return
	}

	script, err := func() (*types.Script, error) {
		bridgeInterfaceIP, err := utils.GetInterfaceIP(e.BridgeInterface, e.BridgePeerCIDR.IsIPv4())
		if err != nil {
			return nil, err
		}
		tunnelInterfaceIP, err := utils.GetInterfaceIP(e.Interface, e.BridgePeerCIDR.IsIPv4())
		if err != nil {
			return nil, err
		}
		return ex.render(&ex.interfaceActivate, map[string]string{
			placeholderBridgeInterface:   e.BridgeInterface,
			placeholderBridgeInterfaceIP: bridgeInterfaceIP,
			placeholderBridgePeerCIDR:    string(e.BridgePeerCIDR),
			placeholderTunnelInterface:   e.Interface,
			placeholderTunnelInterfaceIP: tunnelInterfaceIP,
		}), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
			zap.String("bridge_uuid", ex.bridgeUUID.String()),
		)
		return
	}

	ex.schedule(script)
}

func (ex *Executor) ExecuteInterfaceDeactivate(ctx context.Context, e *event.TunnelInterfaceDeactivated) {
	l := logutils.LoggerFromContext(ctx)

	if len(ex.interfaceDeactivate) == 0 {
		l.Debug("No interface deactivation script configured; skipping...")
		return
	}

	script, err := func() (*types.Script, error) {
		bridgeInterfaceIP, err := utils.GetInterfaceIP(e.BridgeInterface, e.BridgePeerCIDR.IsIPv4())
		if err != nil {
			return nil, err
		}
		tunnelInterfaceIP, err := utils.GetInterfaceIP(e.Interface, e.BridgePeerCIDR.IsIPv4())
		if err != nil {
			return nil, err
		}
		return ex.render(&ex.interfaceDeactivate, map[string]string{
			placeholderBridgeInterface:   e.BridgeInterface,
			placeholderBridgeInterfaceIP: bridgeInterfaceIP,
			placeholderBridgePeerCIDR:    string(e.BridgePeerCIDR),
			placeholderTunnelInterface:   e.Interface,
			placeholderTunnelInterfaceIP: tunnelInterfaceIP,
		}), nil
	}()
	if err != nil {
		l.Error("Failed to execute interface activation script",
			zap.Error(err),
			zap.String("bridge_interface", e.BridgeInterface),
			zap.String("bridge_name", ex.bridgeName),
			zap.String("bridge_uuid", ex.bridgeUUID.String()),
		)
		return
	}

	ex.schedule(script)
}
