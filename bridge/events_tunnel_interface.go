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

	l.Info("Tunnel interface went down",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", e.EvtTunnelInterface()),
	)

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
	s.events <- &event.TunnelInterfaceDeactivated{ // emit event
		BridgeInterface: s.cfg.BridgeInterface,
		BridgePeerCIDR:  s.cfg.PeerCIDR,
		TunnelInterface: e.EvtTunnelInterface(),
		Timestamp:       e.Timestamp,
	}

	// then activate another tunnel

	for promotedIfsName, promotedIfs := range s.status.Interfaces {
		if promotedIfsName == e.EvtTunnelInterface() || !promotedIfs.Up {
			continue
		}
		promotedIfs.Active = true
		s.events <- &event.TunnelInterfaceActivated{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePeerCIDR:  s.cfg.PeerCIDR,
			TunnelInterface: promotedIfsName,
			Timestamp:       e.Timestamp,
		}
		return
	}
}

func (s *Server) eventTunnelInterfaceWentUp(ctx context.Context, e *event.TunnelInterfaceWentUp, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface went up",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", e.EvtTunnelInterface()),
	)

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
				s.events <- &event.TunnelInterfaceDeactivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDR:  s.cfg.PeerCIDR,
					TunnelInterface: demotedIfsName,
					Timestamp:       e.Timestamp,
				}
			}

			// then activate self

			ifs.Active = true
			s.events <- &event.TunnelInterfaceActivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDR:  s.cfg.PeerCIDR,
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
			s.events <- &event.TunnelInterfaceActivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDR:  s.cfg.PeerCIDR,
				TunnelInterface: e.EvtTunnelInterface(),
				Timestamp:       e.Timestamp,
			}
		}
	}
}

func (s *Server) eventTunnelInterfaceDeactivated(ctx context.Context, e *event.TunnelInterfaceDeactivated, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface deactivated",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", e.EvtTunnelInterface()),
	)

	s.executor.ExecuteInterfaceDeactivate(ctx, e)
}

func (s *Server) eventTunnelInterfaceActivated(ctx context.Context, e *event.TunnelInterfaceActivated, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Tunnel interface activated",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", e.EvtTunnelInterface()),
	)

	s.executor.ExecuteInterfaceActivate(ctx, e)
}
