package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/logutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Metrics

	server *http.Server
}

func NewServer(ctx context.Context, cfg *config.Metrics) *Server {
	l := logutils.LoggerFromContext(ctx)

	s := &Server{
		cfg: cfg,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(s.healthcheck))
	mux.Handle("/metrics", promhttp.Handler())

	s.server = &http.Server{
		Addr:              cfg.ListenAddr.String(),
		ErrorLog:          logutils.NewHttpServerErrorLogger(l),
		Handler:           mux,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return s
}

func (s *Server) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	go func() {
		l.Info("VPN HA-monitor metrics server is going up...",
			zap.String("metrics_listen_address", s.server.Addr),
		)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failureSink <- err
		}
		l.Info("VPN HA-monitor metrics server is down")
	}()
}

func (s *Server) Stop(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		l.Error("VPN HA-monitor metrics server shutdown failed",
			zap.Error(err),
		)
	}
}

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
