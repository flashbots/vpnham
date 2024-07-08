package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/transponder"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

var (
	errBridgeReturnProbeDstUUIDMismatch = errors.New("return probe has destination uuid mismatch")
	errBridgeReturnProbeSrcUUIDMismatch = errors.New("return probe has source uuid mismatch")
)

func (s *Server) handleStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	l := logutils.LoggerFromRequest(r)

	defer r.Body.Close()
	if _, err := io.ReadAll(r.Body); err != nil {
		l.Warn("Failed to read request body",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
	}

	if r.Method != http.MethodGet {
		l.Warn("Unexpected status request method",
			zap.String("bridge_name", s.cfg.Name),
			zap.String("method", r.Method),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(s.status); err != nil {
		l.Warn("Failed to encode and send response body",
			zap.Error(err),
			zap.String("bridge_name", s.cfg.Name),
		)
	}
}

func (s *Server) handleProbe(
	ctx context.Context,
	tp *transponder.Transponder,
	from *net.UDPAddr,
	probe *types.Probe,
) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Received probe",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("src_uuid", probe.SrcUUID.String()),
		zap.String("tunnel_interface", tp.InterfaceName()),
		zap.Time("dst_timestamp", probe.DstTimestamp),
		zap.Uint64("sequence", probe.Sequence),
	)

	switch {
	// reply to the other side's probes
	case probe.DstTimestamp.IsZero():
		s.respondToProbe(ctx, tp, from, probe)

	// handle our own returned probes
	case probe.SrcUUID == s.uuid:
		s.processReturnedProbe(ctx, tp, from, probe)

	// handle mismatched probes
	default:
		s.processMismatchedReturnedProbes(ctx, tp, from, probe)
	}
}

func (s *Server) handleTick(
	ctx context.Context,
	ts time.Time,
	failureSink chan<- error,
) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Handling timer tick",
		zap.String("bridge_name", s.cfg.Name),
		zap.Time("tick", ts),
	)

	s.sendProbes(ctx, failureSink)
	s.detectMissedProbes(ctx, failureSink)
	s.pollPartnerBridge(ctx, failureSink)
}
