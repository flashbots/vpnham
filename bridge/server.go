package bridge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/executor"
	"github.com/flashbots/vpnham/httplogger"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/monitor"
	"github.com/flashbots/vpnham/transponder"
	"github.com/flashbots/vpnham/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Bridge

	uuid uuid.UUID

	server   *http.Server
	ticker   *time.Ticker
	executor *executor.Executor

	partner        *types.Partner
	partnerMonitor *monitor.Monitor

	monitors     map[string]*monitor.Monitor
	peers        map[string]*types.Peer
	transponders map[string]*transponder.Transponder

	events chan event.Event

	partnerStatus   *types.BridgeStatus
	mxPartnerStatus sync.Mutex

	status   *types.BridgeStatus
	mxStatus sync.Mutex
}

const (
	pathStatus = "status"
)

func NewServer(ctx context.Context, cfg *config.Bridge) (*Server, error) {
	l := logutils.LoggerFromContext(ctx)

	_uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	cfg.UUID = _uuid

	executor, err := executor.New(cfg)
	if err != nil {
		return nil, err
	}

	partner, err := types.NewPartner(cfg.PartnerURL)
	if err != nil {
		return nil, err
	}

	partnerMonitor, err := monitor.New(cfg.PartnerStatusThresholdDown, cfg.PartnerStatusThresholdUp)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg: cfg,

		uuid: _uuid,

		ticker:   time.NewTicker(cfg.ProbeInterval),
		executor: executor,

		partner:        partner,
		partnerMonitor: partnerMonitor,

		monitors:     make(map[string]*monitor.Monitor, cfg.TunnelInterfacesCount()),
		peers:        make(map[string]*types.Peer, cfg.TunnelInterfacesCount()),
		transponders: make(map[string]*transponder.Transponder, cfg.TunnelInterfacesCount()),

		events: make(chan event.Event, 2*cfg.TunnelInterfacesCount()),

		status: &types.BridgeStatus{
			Name:       cfg.Name,
			Active:     false, // inactive at start, activate only when tunnels are up
			Role:       cfg.Role,
			Interfaces: make(map[string]*types.TunnelInterfaceStatus, cfg.TunnelInterfacesCount()),
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/"+pathStatus, http.HandlerFunc(s.handleStatus))
	handler := httplogger.Middleware(l, mux)

	s.server = &http.Server{
		Addr:              string(cfg.StatusAddr),
		ErrorLog:          logutils.NewHttpServerErrorLogger(l),
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	for ifsName, ifs := range cfg.TunnelInterfaces {
		// monitor
		m, err := monitor.New(ifs.ThresholdDown, ifs.ThresholdUp)
		if err != nil {
			return nil, fmt.Errorf("%s: %w",
				ifsName, err,
			)
		}
		s.monitors[ifsName] = m

		// peer
		peer, err := types.NewPeer(ifsName, ifs.ProbeAddr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w",
				ifsName, err,
			)
		}
		s.peers[ifsName] = peer

		// transponder
		tp, err := transponder.New(ifsName, ifs.Addr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w",
				ifsName, err,
			)
		}
		tp.Receive = s.handleProbe
		s.transponders[ifsName] = tp

		// status
		s.status.Interfaces[ifsName] = &types.TunnelInterfaceStatus{
			Active: false, // inactive at start, activate only when probes report Ok
			Up:     false,
		}
	}

	return s, nil
}

func (s *Server) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	s.executor.Run(ctx, failureSink)

	s.runEventLoop(ctx, failureSink)

	for _, tp := range s.transponders {
		tp.Run(ctx, failureSink)
	}

	go func() {
		l.Info("VPN HA-monitor bridge server is going up...",
			zap.String("bridge_listen_address", s.server.Addr),
			zap.String("bridge_name", s.cfg.Name),
		)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failureSink <- err
		}
		l.Info("VPN HA-monitor bridge server is down",
			zap.String("bridge_name", s.cfg.Name),
		)
	}()

	go func() {
		for {
			s.handleTick(ctx, <-s.ticker.C, failureSink)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx)

	s.executor.Stop(ctx)

	s.stopEventLoop(ctx)

	s.ticker.Stop()

	for _, t := range s.transponders {
		t.Stop(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		l.Error("VPN HA-monitor bridge server shutdown failed",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
	}
}
