package bridge

import (
	"context"
	"time"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

func (s *Server) eventBridgeWentDown(ctx context.Context, e *event.BridgeWentDown, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Bridge going down...")

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()
	s.mxPartnerStatus.Lock()
	defer s.mxPartnerStatus.Unlock()

	if s.status.Active {
		s.status.Active = false
		s.status.ActiveSince = e.Timestamp
		s.events <- &event.BridgeDeactivated{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
			Timestamp:       e.Timestamp,
		}
	}

	if s.partnerStatus == nil || !s.partnerStatus.Up {
		s.events <- &event.ConnectivityLost{ // emit event
			Timestamp: e.Timestamp,
		}
	}
}

func (s *Server) eventBridgeWentUp(ctx context.Context, e *event.BridgeWentUp, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Bridge going up...")

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()
	s.mxPartnerStatus.Lock()
	defer s.mxPartnerStatus.Unlock()

	if !s.status.Active {
		switch s.cfg.Role {
		case types.RoleActive:
			s.status.Active = true
			s.status.ActiveSince = e.Timestamp
			s.events <- &event.BridgeActivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
				Timestamp:       e.Timestamp,
			}

		case types.RoleStandby:
			if s.partnerStatus == nil || !s.partnerStatus.Up {
				s.status.Active = true
				s.status.ActiveSince = e.Timestamp
				s.events <- &event.BridgeActivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
					Timestamp:       e.Timestamp,
				}
			}
		}
	}

	if s.partnerStatus == nil || !s.partnerStatus.Up {
		s.events <- &event.ConnectivityRestored{ // emit event
			Timestamp: e.Timestamp,
		}
	}
}

func (s *Server) eventBridgeDeactivated(ctx context.Context, _ *event.BridgeDeactivated, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Bridge deactivated")

	if r := s.cfg.Reconcile.BridgeActivate.Reapply; r.Enabled() {
		reapply := s.reapply.bridgeActivate
		reapply.Count = 0
		reapply.Next = time.Time{} // disable re-activations of an inactive bridge
	}
}

func (s *Server) eventBridgeActivated(ctx context.Context, e *event.BridgeActivated, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Bridge activating...")

	s.reconciler.BridgeActivate(ctx, e, failureSink)

	if r := s.cfg.Reconcile.BridgeActivate.Reapply; r.Enabled() {
		reapply := s.reapply.bridgeActivate
		reapply.Count = 0
		reapply.Next = e.Timestamp.Add(r.InitialDelay)
	}
}

func (s *Server) eventBridgeReactivated(ctx context.Context, e *event.BridgeReactivated, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	s.mxStatus.Lock()
	if !s.status.Active {
		l.Info("Skipping bridge reactivation since it's already inactive by now")
		s.mxStatus.Unlock()
		return
	}
	defer s.mxStatus.Unlock()

	l.Info("Bridge reactivating...",
		zap.Int("iteration", e.Iteration),
	)

	s.reconciler.BridgeActivate(ctx, e, failureSink)

	if r := s.cfg.Reconcile.BridgeActivate.Reapply; r.Enabled() {
		reapply := s.reapply.bridgeActivate
		reapply.Count++
		reapply.Next = e.Timestamp.Add(r.DelayOnIteration(reapply.Count))
	}
}
