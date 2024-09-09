package bridge

import (
	"context"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

func (s *Server) eventTunnelInterfaceWentDown(ctx context.Context, e *event.TunnelInterfaceWentDown, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface going down...")

	s.deriveBridgeEvents(e)

	//
	// if this tunnel was `active`` when ti went `down`, try to find another
	// tunnel that is `up`` and promote it to be `active`
	//

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	ifs := s.status.Interfaces[e.EvtTunnelInterface()]
	if !ifs.Active {
		return
	}

	// first deactivate self

	ifs.Active = false
	ifs.ActiveSince = e.Timestamp
	s.events <- &event.TunnelInterfaceDeactivated{ // emit event
		BridgeInterface: s.cfg.BridgeInterface,
		BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
		TunnelInterface: e.EvtTunnelInterface(),
		Timestamp:       e.Timestamp,
	}

	// then activate another tunnel

	for promotedIfsName, promotedIfs := range s.status.Interfaces {
		if promotedIfsName == e.EvtTunnelInterface() || !promotedIfs.Up {
			continue
		}
		promotedIfs.Active = true
		promotedIfs.ActiveSince = e.Timestamp
		s.events <- &event.TunnelInterfaceActivated{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
			TunnelInterface: promotedIfsName,
			Timestamp:       e.Timestamp,
		}
		return
	}
}

func (s *Server) eventTunnelInterfaceWentUp(ctx context.Context, e *event.TunnelInterfaceWentUp, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface going up...")

	s.deriveBridgeEvents(e)

	//
	// when going up:
	//
	//   - if this tunnel is configured `active`, overtake the active status
	//
	//   - otherwise, only elect self to be `active` if there's no other active
	//     tunnel around
	//

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	ifs := s.status.Interfaces[e.EvtTunnelInterface()]
	cfg := s.cfg.TunnelInterfaces[e.EvtTunnelInterface()]
	switch cfg.Role {
	case types.RoleActive:
		if !ifs.Active {
			// first deactivate other tunnel (if needed)

			for demotedIfsName, demotedIfs := range s.status.Interfaces {
				if demotedIfsName == e.EvtTunnelInterface() || !demotedIfs.Active {
					continue
				}
				demotedIfs.Active = false
				demotedIfs.ActiveSince = e.Timestamp
				s.events <- &event.TunnelInterfaceDeactivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
					TunnelInterface: demotedIfsName,
					Timestamp:       e.Timestamp,
				}
			}

			// then activate self

			ifs.Active = true
			ifs.ActiveSince = e.Timestamp
			s.events <- &event.TunnelInterfaceActivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
				TunnelInterface: e.EvtTunnelInterface(),
				Timestamp:       e.Timestamp,
			}

		}

	case types.RoleStandby:
		anotherActiveIfsExists := false
		for anotherIfsName, anotherIfs := range s.status.Interfaces {
			if anotherIfsName == e.EvtTunnelInterface() {
				continue
			}
			if anotherIfs.Active {
				anotherActiveIfsExists = true
				break
			}
		}
		if !anotherActiveIfsExists {
			ifs.Active = true
			ifs.ActiveSince = e.Timestamp
			s.events <- &event.TunnelInterfaceActivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
				TunnelInterface: e.EvtTunnelInterface(),
				Timestamp:       e.Timestamp,
			}
		}
	}
}

func (s *Server) eventTunnelInterfaceDeactivated(ctx context.Context, e *event.TunnelInterfaceDeactivated, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface deactivating...")

	s.reconciler.InterfaceDeactivate(ctx, e, failureSink)
}

func (s *Server) eventTunnelInterfaceActivated(ctx context.Context, e *event.TunnelInterfaceActivated, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface activating...")

	s.reconciler.InterfaceActivate(ctx, e, failureSink)

	if r := s.cfg.Reconcile.InterfaceActivate.Reapply; r.Enabled() {
		reapply := s.reapply.interfaceActivate
		reapply.Count = 0
		reapply.Next = e.Timestamp.Add(r.InitialDelay)
	}
}

func (s *Server) eventTunnelInterfaceReactivated(ctx context.Context, e *event.TunnelInterfaceReactivated, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface reactivating...",
		zap.Int("iteration", e.Iteration),
	)

	s.reconciler.InterfaceActivate(ctx, e, failureSink)

	if r := s.cfg.Reconcile.InterfaceActivate.Reapply; r.Enabled() {
		reapply := s.reapply.interfaceActivate
		reapply.Count++
		reapply.Next = e.Timestamp.Add(r.DelayOnIteration(reapply.Count))
	}
}
