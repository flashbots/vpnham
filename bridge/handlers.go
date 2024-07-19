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
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/transponder"
	"github.com/flashbots/vpnham/types"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
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
		l.Error("Failed to read request body",
			zap.Error(err),
		)
		metrics.Errors.Add(r.Context(), 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, s.cfg.Name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeStatusListener),
		))
	}

	if r.Method != http.MethodGet {
		l.Error("Unexpected status request method",
			zap.String("method", r.Method),
		)
		metrics.Errors.Add(r.Context(), 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, s.cfg.Name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeStatusListener),
		))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.mxStatus.Lock()
	defer s.mxStatus.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(s.status); err != nil {
		l.Error("Failed to encode and send response body",
			zap.Error(err),
		)
		metrics.Errors.Add(r.Context(), 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelBridge, s.cfg.Name),
			attribute.String(metrics.LabelErrorScope, metrics.ScopeStatusListener),
		))
		return
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

	l.Debug("Handling bridge timer tick",
		zap.Time("tick", ts),
	)

	s.sendProbes(ctx, failureSink)
	s.detectMissedProbes(ctx, failureSink)
	s.pollPartnerBridge(ctx, failureSink)
	s.reapplyUpdates(ctx, failureSink)
}
