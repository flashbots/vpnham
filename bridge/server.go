package bridge

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/httplogger"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/monitor"
	"github.com/flashbots/vpnham/reconciler"
	"github.com/flashbots/vpnham/transponder"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	cfg *config.Bridge

	uuid uuid.UUID

	reconciler *reconciler.Reconciler
	server     *http.Server
	ticker     *time.Ticker

	http           *http.Client
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

	reapply struct {
		bridgeActivate    *types.ReapplyStatus
		interfaceActivate *types.ReapplyStatus
	}
}

const (
	pathStatus = "status"
)

func NewServer(ctx context.Context, cfg *config.Bridge) (*Server, error) {
	l := logutils.LoggerFromContext(ctx)

	ts := time.Now()

	_uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	cfg.UUID = _uuid

	reconciler, err := reconciler.New(cfg.Name, cfg.Reconcile)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   cfg.PartnerStatusTimeout,
		KeepAlive: 2 * cfg.PartnerStatusTimeout,
	}
	if cfg.PartnerPollingInterface != "" {
		ipv4s, ipv6s, err := utils.GetInterfaceIPs(cfg.PartnerPollingInterface)
		if err != nil {
			return nil, err
		}
		if len(ipv4s) > 0 {
			dialer.LocalAddr = &net.TCPAddr{IP: net.ParseIP(ipv4s[0])}
		} else if len(ipv6s) > 0 {
			dialer.LocalAddr = &net.TCPAddr{IP: net.ParseIP(ipv6s[0])}
		}
	}
	transport := &http.Transport{
		DialContext:     dialer.DialContext,
		IdleConnTimeout: 4 * cfg.PartnerStatusTimeout,
		MaxIdleConns:    2,
	}
	cli := &http.Client{
		Transport: transport,
		Timeout:   cfg.PartnerStatusTimeout,
	}

	partner, err := types.NewPartner(cfg.PartnerURL)
	if err != nil {
		return nil, err
	}

	partnerMonitor, err := func() (*monitor.Monitor, error) {
		if cfg.Role == types.RoleActive {
			return monitor.New(cfg.PartnerStatusThresholdDown, cfg.PartnerStatusThresholdUp)
		} else {
			return monitor.New(cfg.PartnerStatusThresholdDown+1, cfg.PartnerStatusThresholdUp+1)
		}
	}()
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg: cfg,

		uuid: _uuid,

		reconciler: reconciler,
		ticker:     time.NewTicker(cfg.ProbeInterval),

		http:           cli,
		partner:        partner,
		partnerMonitor: partnerMonitor,

		monitors:     make(map[string]*monitor.Monitor, cfg.TunnelInterfacesCount()),
		peers:        make(map[string]*types.Peer, cfg.TunnelInterfacesCount()),
		transponders: make(map[string]*transponder.Transponder, cfg.TunnelInterfacesCount()),

		events: make(chan event.Event, 2*cfg.TunnelInterfacesCount()),

		status: &types.BridgeStatus{
			Name:        cfg.Name,
			Active:      false, // inactive at start, activate only when tunnels are up
			ActiveSince: ts,
			Role:        cfg.Role,
			Interfaces:  make(map[string]*types.TunnelInterfaceStatus, cfg.TunnelInterfacesCount()),
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/"+pathStatus, http.HandlerFunc(s.handleStatus))
	handler := httplogger.Middleware(l, mux)

	s.server = &http.Server{
		Addr:              cfg.StatusAddr.String(),
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
		tp, err := transponder.New(cfg.Name, ifsName, ifs.Addr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w",
				ifsName, err,
			)
		}
		tp.Receive = s.handleProbe
		s.transponders[ifsName] = tp

		// status
		s.status.Interfaces[ifsName] = &types.TunnelInterfaceStatus{
			Active:      false, // inactive at start, activate only when probes report Ok
			ActiveSince: ts,
			Up:          false,
			UpSince:     ts,
		}
	}

	if cfg.Reconcile.BridgeActivate.Reapply.Enabled() {
		s.reapply.bridgeActivate = &types.ReapplyStatus{}
	}
	if cfg.Reconcile.InterfaceActivate.Reapply.Enabled() {
		s.reapply.interfaceActivate = &types.ReapplyStatus{}
	}

	return s, nil
}

func (s *Server) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx).With(
		zap.String("bridge_name", s.cfg.Name),
	)
	ctx = logutils.ContextWithLogger(ctx, l)

	s.reconciler.Run(ctx, failureSink)

	s.runEventLoop(ctx, failureSink)

	for _, tp := range s.transponders {
		tp.Run(ctx, failureSink)
	}

	go func() {
		l.Info("VPN HA-monitor bridge server is going up...",
			zap.String("bridge_listen_address", s.server.Addr),
		)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failureSink <- err
		}
		l.Info("VPN HA-monitor bridge server is down")
	}()

	go func() {
		for {
			s.handleTick(ctx, <-s.ticker.C, failureSink)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx).With(
		zap.String("bridge_name", s.cfg.Name),
	)
	ctx = logutils.ContextWithLogger(ctx, l)

	s.reconciler.Stop(ctx)

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
		)
	}
}
