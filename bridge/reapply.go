package bridge

import (
	"context"
	"time"

	"github.com/flashbots/vpnham/event"
)

func (s *Server) reapplyUpdates(_ context.Context, _ chan<- error) {
	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	if reapply := s.reapply.bridgeActivate; reapply != nil {
		if !reapply.Next.IsZero() && time.Now().After(reapply.Next) {
			if s.status.Active {
				reapply.Next = time.Time{} // avoid re-fire

				s.events <- &event.BridgeReactivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
					Iteration:       reapply.Count,
					Timestamp:       time.Now(),
				}
			}
		}
	}

	if reapply := s.reapply.interfaceActivate; reapply != nil {
		if !reapply.Next.IsZero() && time.Now().After(reapply.Next) {
			if activeInterface := s.status.ActiveInterface(); activeInterface != "" {
				reapply.Next = time.Time{} // avoid re-fire

				s.events <- &event.TunnelInterfaceReactivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDRs: s.cfg.BridgePeerCIDRs(),
					Iteration:       reapply.Count,
					TunnelInterface: activeInterface,
					Timestamp:       time.Now(),
				}
			}
		}
	}
}
