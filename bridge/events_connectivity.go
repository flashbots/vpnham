package bridge

import (
	"context"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
)

func (s *Server) eventConnectivityLost(ctx context.Context, _ *event.ConnectivityLost, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Connectivity lost")
}

func (s *Server) eventConnectivityRestored(ctx context.Context, _ *event.ConnectivityRestored, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Connectivity restored")
}
