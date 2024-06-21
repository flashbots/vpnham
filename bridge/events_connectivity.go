package bridge

import (
	"context"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"go.uber.org/zap"
)

func (s *Server) eventConnectivityLost(ctx context.Context, _ *event.ConnectivityLost, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Connectivity lost",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("bridge_uuid", s.uuid.String()),
	)
}

func (s *Server) eventConnectivityRestored(ctx context.Context, _ *event.ConnectivityRestored, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Connectivity restored",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("bridge_uuid", s.uuid.String()),
	)
}
