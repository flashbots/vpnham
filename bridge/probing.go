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
					zap.String("bridge_uuid", s.uuid.String()),
					zap.String("tunnel_interface", peer.InterfaceName()),
				)
				s.events <- &event.TunnelProbeSendFailure{ // emit event
					Interface: ifsName,
					Sequence:  probe.Sequence,
					Timestamp: probe.SrcTimestamp,
				}
				return
			}

			s.events <- &event.TunnelProbeSendSuccess{ // emit event
				Interface: ifsName,
				Sequence:  probe.Sequence,
				Timestamp: probe.SrcTimestamp,
			}
		})
	}
}

func (s *Server) detectMissedProbes(_ context.Context, _ chan<- error) {
	for ifsName, peer := range s.peers {
		for missed := peer.Acknowledgement() + 1; missed < peer.Sequence(); missed++ {
			s.events <- &event.TunnelProbeReturnFailure{
				Interface: ifsName,
				Timestamp: time.Now(),
				Sequence:  missed,
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

	probe.DstLocation = s.cfg.ProbeLocation
	probe.DstTimestamp = time.Now()

	tp.SendProbe(probe, from, func(err error) {
		if err != nil {
			l.Warn("Failed to respond to incoming probe",
				zap.Error(err),
				zap.String("bridge_name", s.cfg.Name),
				zap.String("bridge_uuid", s.uuid.String()),
				zap.String("probe_src_addr", from.String()),
				zap.String("probe_src_location", probe.SrcLocation.String()),
				zap.String("probe_src_uuid", probe.SrcUUID.String()),
				zap.String("tunnel_interface", tp.InterfaceName()),
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

	peer := s.peers[tp.InterfaceName()]
	// check for errors
	if probe.DstUUID != peer.UUID() {
		l.Warn("Invalid return probe",
			zap.String("bridge_name", s.cfg.Name),
			zap.String("bridge_uuid", s.uuid.String()),
			zap.String("tunnel_interface", tp.InterfaceName()),

			zap.Error(fmt.Errorf("%w: expected %s, got %s",
				errBridgeReturnProbeDstUUIDMismatch, peer.UUID(), probe.DstUUID,
			)),
		)
		return
	}
	// detect missed probes
	for missed := peer.Acknowledgement() + 1; missed < probe.Sequence; missed++ {
		s.events <- &event.TunnelProbeReturnFailure{
			Interface: tp.InterfaceName(),
			Sequence:  missed,
			Timestamp: ts,
		}
	}
	peer.SetAcknowledgement(probe.Sequence)
	s.events <- &event.TunnelProbeReturnSuccess{ // emit event
		Interface:      tp.InterfaceName(),
		LatencyForward: probe.DstTimestamp.Sub(probe.SrcTimestamp),
		LatencyReturn:  ts.Sub(probe.DstTimestamp),
		Location:       probe.DstLocation.String(),
		Sequence:       probe.Sequence,
		Timestamp:      ts,
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
		zap.String("bridge_uuid", s.uuid.String()),
		zap.String("tunnel_interface", tp.InterfaceName()),

		zap.Error(fmt.Errorf("%w: expected %s, got %s",
			errBridgeReturnProbeSrcUUIDMismatch, s.uuid, probe.SrcUUID,
		)),
	)
}
