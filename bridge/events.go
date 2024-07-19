package bridge

import (
	"context"
	"reflect"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/monitor"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (s *Server) runEventLoop(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	go func() {
		l.Info("VPN HA-monitor bridge event loop is starting...")

		for e := range s.events {
			var ( // use fresh originals on every iteration
				ctx = ctx
				l   = l
			)

			if te, ok := e.(event.TunnelInterfaceEvent); ok {
				l = l.With(
					zap.String("tunnel_interface", te.EvtTunnelInterface()),
				)
				ctx = logutils.ContextWithLogger(ctx, l)
			}

			switch e := e.(type) {

			// bridge

			case *event.BridgeActivated:
				s.eventBridgeActivated(ctx, e, failureSink)
			case *event.BridgeDeactivated:
				s.eventBridgeDeactivated(ctx, e, failureSink)
			case *event.BridgeReactivated:
				s.eventBridgeReactivated(ctx, e, failureSink)
			case *event.BridgeWentDown:
				s.eventBridgeWentDown(ctx, e, failureSink)
			case *event.BridgeWentUp:
				s.eventBridgeWentUp(ctx, e, failureSink)

			// connectivity

			case *event.ConnectivityLost:
				s.eventConnectivityLost(ctx, e, failureSink)
			case *event.ConnectivityRestored:
				s.eventConnectivityRestored(ctx, e, failureSink)

			// partner

			case *event.PartnerActivated:
				s.eventPartnerActivated(ctx, e, failureSink)
			case *event.PartnerChangedName:
				s.eventPartnerChangedName(ctx, e, failureSink)
			case *event.PartnerDeactivated:
				s.eventPartnerDeactivated(ctx, e, failureSink)
			case *event.PartnerPollFailure:
				s.eventPartnerPollFailure(ctx, e, failureSink)
			case *event.PartnerPollSuccess:
				s.eventPartnerPollSuccess(ctx, e, failureSink)
			case *event.PartnerWentDown:
				s.eventPartnerWentDown(ctx, e, failureSink)
			case *event.PartnerWentUp:
				s.eventPartnerWentUp(ctx, e, failureSink)

			// tunnel

			case *event.TunnelInterfaceActivated:
				s.eventTunnelInterfaceActivated(ctx, e, failureSink)
			case *event.TunnelInterfaceDeactivated:
				s.eventTunnelInterfaceDeactivated(ctx, e, failureSink)
			case *event.TunnelInterfaceReactivated:
				s.eventTunnelInterfaceReactivated(ctx, e, failureSink)
			case *event.TunnelInterfaceWentDown:
				s.eventTunnelInterfaceWentDown(ctx, e, failureSink)
			case *event.TunnelInterfaceWentUp:
				s.eventTunnelInterfaceWentUp(ctx, e, failureSink)
			case *event.TunnelProbeReturnFailure:
				s.eventTunnelProbeReturnFailure(ctx, e, failureSink)
			case *event.TunnelProbeReturnSuccess:
				s.eventTunnelProbeReturnSuccess(ctx, e, failureSink)
			case *event.TunnelProbeSendFailure:
				s.eventTunnelProbeSendFailure(ctx, e, failureSink)
			case *event.TunnelProbeSendSuccess:
				s.eventTunnelProbeSendSuccess(ctx, e, failureSink)

			// catch-all

			default:
				l.Error("Unexpected event",
					zap.String("kind", e.EvtKind()),
					zap.String("type", reflect.TypeOf(e).String()),
				)
				metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
					attribute.String(metrics.LabelBridge, s.cfg.Name),
					attribute.String(metrics.LabelErrorScope, metrics.ScopeInternalLogic),
				))
			}
		}

		l.Info("VPN HA-monitor bridge event loop is stopped")
	}()
}

func (s *Server) stopEventLoop(_ context.Context) {
	close(s.events)
}

// detectTunnelUpDownEvents derives tunnel up/down events from tunnel probe events
func (s *Server) detectTunnelUpDownEvents(e event.TunnelInterfaceEvent, updateMonitor func(*monitor.Monitor)) {
	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	ifs := s.status.Interfaces[e.EvtTunnelInterface()]
	mon := s.monitors[e.EvtTunnelInterface()]

	updateMonitor(mon)

	switch mon.Status() {
	case monitor.Down:
		if ifs.Up {
			ifs.Up = false
			ifs.UpSince = e.EvtTimestamp()
			s.events <- &event.TunnelInterfaceWentDown{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDR:  s.cfg.PeerCIDR,
				TunnelInterface: e.EvtTunnelInterface(),
				Timestamp:       e.EvtTimestamp(),
			}
		}

	case monitor.Up:
		if !ifs.Up {
			ifs.Up = true
			ifs.UpSince = e.EvtTimestamp()
			s.events <- &event.TunnelInterfaceWentUp{ // emit event
				BridgeInterface: s.cfg.BridgeInterface,
				BridgePeerCIDR:  s.cfg.PeerCIDR,
				TunnelInterface: e.EvtTunnelInterface(),
				Timestamp:       e.EvtTimestamp(),
			}
		}
	}
}

// deriveBridgeEvents derives bridge events from tunnel-interface events
func (s *Server) deriveBridgeEvents(e event.TunnelInterfaceEvent) {
	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	up := false
	for _, ifs := range s.status.Interfaces {
		up = up || ifs.Up
	}

	if up == s.status.Up {
		return
	}

	if up {
		s.events <- &event.BridgeWentUp{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePerCIDR:   s.cfg.PeerCIDR,
			Timestamp:       e.EvtTimestamp(),
		}
	} else {
		s.events <- &event.BridgeWentDown{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePeerCIDR:  s.cfg.PeerCIDR,
			Timestamp:       e.EvtTimestamp(),
		}
	}

	s.status.Up = up
}

// derivePartnerUpDownEvents derives partner up/down events from partner poll events
func (s *Server) derivePartnerUpDownEvents(e event.PartnerPollEvent, updateMonitor func(*monitor.Monitor)) {
	s.mxPartnerStatus.Lock()
	defer s.mxPartnerStatus.Unlock()

	firstContact := false
	if s.partnerStatus == nil {
		if firstStatus := e.PartnerStatus(); firstStatus != nil {
			firstContact = true
			s.partnerStatus = firstStatus
		} else {
			// old status is nil, new status is also nil => nothing to do here
			return
		}
	}

	updateMonitor(s.partnerMonitor)

	switch s.partnerMonitor.Status() {
	case monitor.Down:
		if s.partnerStatus.Up {
			s.partnerStatus.Up = false
			s.events <- &event.PartnerWentDown{ // emit event
				Timestamp: e.EvtTimestamp(),
			}
		}

	case monitor.Up:
		if firstContact || !s.partnerStatus.Up {
			s.partnerStatus.Up = true
			s.events <- &event.PartnerWentUp{ // emit events
				Timestamp: e.EvtTimestamp(),
			}
		}
	}

	// sync the rest of status

	if newPartnerStatus := e.PartnerStatus(); newPartnerStatus != nil {
		s.partnerStatus.Interfaces = newPartnerStatus.Interfaces

		if s.partnerStatus.Name != newPartnerStatus.Name {
			s.events <- &event.PartnerChangedName{ // emit event
				OldName:   s.partnerStatus.Name,
				NewName:   newPartnerStatus.Name,
				Timestamp: e.EvtTimestamp(),
			}
			s.partnerStatus.Name = newPartnerStatus.Name
		}

		if s.partnerStatus.Active != newPartnerStatus.Active {
			if newPartnerStatus.Active {
				s.events <- &event.PartnerActivated{ // emit event
					Timestamp: e.EvtTimestamp(),
				}
			} else {
				s.events <- &event.PartnerDeactivated{ // emit event
					Timestamp: e.EvtTimestamp(),
				}
			}
			s.partnerStatus.Active = newPartnerStatus.Active
			s.partnerStatus.ActiveSince = newPartnerStatus.ActiveSince
		}
	}
}
