package bridge

import (
	"context"
	"errors"
	"fmt"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/monitor"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	errParterChangedName = errors.New("partner's bridge name has changed")
)

func (s *Server) eventPartnerPollFailure(_ context.Context, e *event.PartnerPollFailure, _ chan<- error) {
	s.derivePartnerUpDownEvents(e, func(m *monitor.Monitor) {
		m.RegisterStatus(e.Sequence, monitor.Down)
	})
}

func (s *Server) eventPartnerChangedName(_ context.Context, e *event.PartnerChangedName, failureSink chan<- error) {
	failureSink <- fmt.Errorf("%w: was %s, is %s",
		errParterChangedName, e.OldName, e.NewName,
	)
}

func (s *Server) eventPartnerPollSuccess(_ context.Context, e *event.PartnerPollSuccess, failureSink chan<- error) {
	if e.Status.Name != s.cfg.Name {
		failureSink <- fmt.Errorf("%w: expected %s, got %s",
			errPartnerBridgeNameIsDifferent, s.cfg.Name, e.Status.Role,
		)
		return
	}

	if e.Status.Role == s.cfg.Role {
		failureSink <- fmt.Errorf("%w: %s",
			errPartnerRoleIsIdentical, e.Status.Role,
		)
		return
	}

	s.derivePartnerUpDownEvents(e, func(m *monitor.Monitor) {
		if e.Status.Up {
			m.RegisterStatus(e.Sequence, monitor.Up)
		} else {
			m.RegisterStatus(e.Sequence, monitor.Down)
		}
	})
}

func (s *Server) eventPartnerWentDown(ctx context.Context, e *event.PartnerWentDown, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Partner went down",
		zap.String("bridge_name", s.cfg.Name),
	)

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()
	s.mxPartnerStatus.Lock()
	defer s.mxPartnerStatus.Unlock()

	if s.partnerStatus.Active {
		s.partnerStatus.Active = false
		s.partnerStatus.ActiveSince = e.EvtTimestamp()
		s.events <- &event.PartnerDeactivated{ // emit event
			Timestamp: e.EvtTimestamp(),
		}
	}

	if !s.status.Active {
		s.status.Active = true
		s.status.ActiveSince = e.Timestamp
		s.events <- &event.BridgeActivated{ // emit event
			BridgeInterface: s.cfg.BridgeInterface,
			BridgePeerCIDR:  s.cfg.PeerCIDR,
			Timestamp:       e.Timestamp,
		}
	}
}

func (s *Server) eventPartnerWentUp(ctx context.Context, e *event.PartnerWentUp, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Partner went up",
		zap.String("bridge_name", s.cfg.Name),
	)

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	if s.status.Active && s.cfg.Role != types.RoleActive {
		s.status.Active = false
		s.status.ActiveSince = e.Timestamp
		s.events <- &event.BridgeDeactivated{
			Timestamp: e.Timestamp,
		}
	}
}

func (s *Server) eventPartnerDeactivated(ctx context.Context, _ *event.PartnerDeactivated, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Partner deactivated",
		zap.String("bridge_name", s.cfg.Name),
	)
}

func (s *Server) eventPartnerActivated(ctx context.Context, _ *event.PartnerActivated, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	logFields := []zapcore.Field{
		zap.String("bridge_name", s.cfg.Name),
	}
	if s.partnerStatus != nil {
		logFields = append(logFields, zap.String("partner_name", s.partnerStatus.Name))
	}
	l.Info("Partner activated", logFields...)
}
