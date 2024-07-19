package bridge

import (
	"context"
	"time"

	"github.com/flashbots/vpnham/event"
)

func (s *Server) reapplyUpdates(_ context.Context, _ chan<- error) {
	if ba := s.reapply.bridgeActivate; ba != nil {
		if !ba.Next.IsZero() && time.Now().After(ba.Next) {
			ba.Next = time.Time{} // avoid re-fire

			s.events <- &event.BridgeReactivated{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDR:  s.cfg.PeerCIDR,
				Iteration:       ba.Count,
				Timestamp:       time.Now(),
			}
		}
	}

	if ia := s.reapply.interfaceActivate; ia != nil {
		if !ia.Next.IsZero() && time.Now().After(ia.Next) {
			ia.Next = time.Time{} // avoid re-fire

			s.mxStatus.Lock()
			defer s.mxStatus.Unlock()

			if activeInterface := s.status.ActiveInterface(); activeInterface != "" {
				s.events <- &event.TunnelInterfaceReactivated{ // emit event
					BridgeInterface: s.cfg.BridgeInterface,
					BridgePeerCIDR:  s.cfg.PeerCIDR,
					Iteration:       ia.Count,
					TunnelInterface: activeInterface,
					Timestamp:       time.Now(),
				}
			}
		}
	}
}
