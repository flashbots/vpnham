package bridge

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/transponder"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

func (s *Server) sendProbes(ctx context.Context, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	for ifsName, peer := range s.peers {
		l.Debug("Sending probe to a peer...",
			zap.String("tunnel_interface", peer.InterfaceName()),
		)

		probe := &types.Probe{
			Sequence:     peer.NextSequence(),
			SrcUUID:      s.uuid,
			SrcLocation:  s.cfg.ProbeLocation,
			SrcTimestamp: time.Now(),
			DstUUID:      peer.UUID(),
		}

		s.transponders[ifsName].SendProbe(probe, peer.ProbeAddr(), func(err error) {
			if err != nil {
				l.Warn("Failed to send a probe",
					zap.Error(err),
					zap.String("bridge_name", s.cfg.Name),
					zap.String("tunnel_interface", peer.InterfaceName()),
					zap.Uint64("sequence", probe.Sequence),
				)
				s.events <- &event.TunnelProbeSendFailure{ // emit event
					TunnelInterface: ifsName,
					ProbeSequence:   probe.Sequence,
					Timestamp:       probe.SrcTimestamp,
				}
				return
			}

			l.Debug("Sent a probe",
				zap.String("bridge_name", s.cfg.Name),
				zap.String("tunnel_interface", peer.InterfaceName()),
				zap.Uint64("sequence", probe.Sequence),
			)
			s.events <- &event.TunnelProbeSendSuccess{ // emit event
				TunnelInterface: ifsName,
				ProbeSequence:   probe.Sequence,
				Timestamp:       probe.SrcTimestamp,
			}
		})
	}
}

func (s *Server) detectMissedProbes(ctx context.Context, _ chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	for ifsName, peer := range s.peers {
		for missed := peer.Acknowledgement() + 1; missed < peer.Sequence(); missed++ {
			l.Debug("Missed a probe (gap in acknowledgement)",
				zap.String("bridge_name", s.cfg.Name),
				zap.String("tunnel_interface", peer.InterfaceName()),
				zap.Uint64("sequence", missed),
			)
			s.events <- &event.TunnelProbeReturnFailure{ // emit event
				TunnelInterface: ifsName,
				Timestamp:       time.Now(),
				ProbeSequence:   missed,
			}
		}
		peer.SetAcknowledgement(peer.Sequence() - 1)
	}
}

func (s *Server) respondToProbe(
	ctx context.Context,
	tp *transponder.Transponder,
	from *net.UDPAddr,
	probe *types.Probe,
) {
	l := logutils.LoggerFromContext(ctx)

	// we only fill our location and our timestamp;  the uuid is filled
	// in by the sender (for the sender's bookkeeping)
	probe.DstLocation = s.cfg.ProbeLocation
	probe.DstTimestamp = time.Now()

	tp.SendProbe(probe, from, func(err error) {
		if err == nil {
			l.Debug("Responded to probe",
				zap.String("bridge_name", s.cfg.Name),
				zap.String("dst_uuid", probe.DstUUID.String()),
				zap.String("tunnel_interface", tp.InterfaceName()),
				zap.Time("dst_timestamp", probe.DstTimestamp),
				zap.Uint64("sequence", probe.Sequence),
			)
		} else {
			l.Warn("Failed to respond to incoming probe",
				zap.Error(err),
				zap.String("bridge_name", s.cfg.Name),
				zap.String("tunnel_interface", tp.InterfaceName()),
				zap.Uint64("sequence", probe.Sequence),
			)
		}
	})
}

func (s *Server) processReturnedProbe(
	ctx context.Context,
	tp *transponder.Transponder,
	_ *net.UDPAddr,
	probe *types.Probe,
) {
	ts := time.Now()

	l := logutils.LoggerFromContext(ctx)

	l.Debug("Processing returned probe...",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", tp.InterfaceName()),
		zap.Uint64("sequence", probe.Sequence),
	)

	peer := s.peers[tp.InterfaceName()]
	// check for errors
	if probe.DstUUID != peer.UUID() {
		l.Warn("Invalid return probe",
			zap.String("bridge_name", s.cfg.Name),
			zap.String("tunnel_interface", tp.InterfaceName()),

			zap.Error(fmt.Errorf("%w: expected %s, got %s",
				errBridgeReturnProbeDstUUIDMismatch, peer.UUID(), probe.DstUUID,
			)),
		)
		return
	}
	// detect missed probes
	for missed := peer.Acknowledgement() + 1; missed < probe.Sequence; missed++ {
		l.Debug("Missed a probe (later probe came in)",
			zap.String("bridge_name", s.cfg.Name),
			zap.String("tunnel_interface", peer.InterfaceName()),
			zap.Uint64("sequence", missed),
		)
		s.events <- &event.TunnelProbeReturnFailure{ // emit event
			TunnelInterface: tp.InterfaceName(),
			ProbeSequence:   missed,
			Timestamp:       ts,
		}
	}
	peer.SetAcknowledgement(probe.Sequence)
	s.events <- &event.TunnelProbeReturnSuccess{ // emit event
		TunnelInterface: tp.InterfaceName(),
		LatencyForward:  probe.DstTimestamp.Sub(probe.SrcTimestamp),
		LatencyReturn:   ts.Sub(probe.DstTimestamp),
		Location:        probe.DstLocation.String(),
		ProbeSequence:   probe.Sequence,
		Timestamp:       ts,
	}
}

func (s *Server) processMismatchedReturnedProbes(
	ctx context.Context,
	tp *transponder.Transponder,
	_ *net.UDPAddr,
	probe *types.Probe,
) {
	l := logutils.LoggerFromContext(ctx)

	l.Warn("Invalid probe",
		zap.String("bridge_name", s.cfg.Name),
		zap.String("tunnel_interface", tp.InterfaceName()),

		zap.Error(fmt.Errorf("%w: expected %s, got %s",
			errBridgeReturnProbeSrcUUIDMismatch, s.uuid, probe.SrcUUID,
		)),
	)
}
