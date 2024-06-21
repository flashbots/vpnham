package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/flashbots/vpnham/bridge"
	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Config
	log *zap.Logger

	bridges map[string]*bridge.Server
	metrics *metrics.Server
}

func New(cfg *config.Config) (*Server, error) {
	l := zap.L()
	ctx := logutils.ContextWithLogger(context.Background(), l)

	bridges := make(map[string]*bridge.Server, len(cfg.Server.Bridges))
	for bn, b := range cfg.Server.Bridges {
		bs, err := bridge.NewServer(ctx, b)
		if err != nil {
			return nil, err
		}
		bridges[bn] = bs
	}

	srv := &Server{
		cfg: cfg,
		log: l,

		bridges: bridges,
	}

	if err := metrics.Setup(ctx, cfg.Server.Metrics, srv.observeMetrics); err != nil {
		return nil, err
	}
	srv.metrics = metrics.NewServer(ctx, cfg.Server.Metrics)

	return srv, nil
}

func (s *Server) Run() error {
	l := s.log
	ctx := logutils.ContextWithLogger(context.Background(), l)

	errs := []error{}
	failureSink := make(chan error, s.cfg.Server.EventSourcesCount())
	defer close(failureSink)

	s.metrics.Run(ctx, failureSink)
	for _, b := range s.bridges {
		b.Run(ctx, failureSink)
	}

	{ // wait until termination or internal failure
		terminator := make(chan os.Signal, 1)
		defer close(terminator)

		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)

		select {
		case stop := <-terminator:
			l.Info("Stop signal received; shutting down...",
				zap.String("signal", stop.String()),
			)
		case err := <-failureSink:
			l.Error("Internal failure; shutting down...",
				zap.Error(err),
			)
			errs = append(errs, err)
		readErrors:
			for { // exhaust the errors
				select {
				case err := <-failureSink:
					l.Error("Extra internal failure",
						zap.Error(err),
					)
					errs = append(errs, err)
				default:
					break readErrors
				}
			}
		}
	}

	for _, bridge := range s.bridges {
		bridge.Stop(ctx)
	}
	s.metrics.Stop(ctx)

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *Server) observeMetrics(ctx context.Context, observer otelapi.Observer) error {
	errs := []error{}

	for bridgeName, bridge := range s.bridges {
		if err := bridge.ObserveMetrics(ctx, observer); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w",
				bridgeName, err,
			))
		}
	}

	switch len(errs) {
	default:
		return errors.Join(errs...)
	case 1:
		return errs[0]
	case 0:
		return nil
	}
}
