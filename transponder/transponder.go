package transponder

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

type Transponder struct {
	ifsAddr *net.UDPAddr
	ifsName string

	conn      *net.UDPConn
	goingDown bool

	Receive Receiver
}

type Receiver = func(context.Context, *Transponder, *net.UDPAddr, *types.Probe)

var (
	errTransponderConnectionIsNotOpen = errors.New("transponder connection is not open")
	errTransponderIsAlreadyGoingDown  = errors.New("transponder is already going down")
	errTransponderIsAlreadyConnected  = errors.New("transponder is already connected")
	errTransponderIsNotConnected      = errors.New("transponder is not connected yet")
	errTransponderReceiverIsNotSet    = errors.New("transponder receiver method is not set")
)

func New(ifsName string, ifsAddr types.Address) (*Transponder, error) {
	ip, port, err := ifsAddr.Parse()
	if err != nil {
		return nil, err
	}

	return &Transponder{
		ifsName: ifsName,
		ifsAddr: &net.UDPAddr{IP: ip, Port: port},
	}, nil
}

func (tp *Transponder) InterfaceName() string {
	return tp.ifsName
}

func (tp *Transponder) Run(ctx context.Context, failureSink chan<- error) {
	if tp.goingDown {
		return
	}

	if err := tp.connect(); err != nil {
		failureSink <- fmt.Errorf("%s: %w", tp.ifsAddr, err)
		return
	}

	go tp.listen(ctx, failureSink)
}

func (tp *Transponder) Stop(ctx context.Context) {
	l := logutils.LoggerFromContext(ctx)

	tp.goingDown = true

	if tp.conn == nil {
		return
	}

	if err := tp.disconnect(); err != nil {
		l.Error("VPN HA-monitor transponder shutdown failed",
			zap.Error(err),
			zap.String("transponder_address", tp.ifsAddr.String()),
		)
	}
	l.Info("VPN HA-monitor transponder is down",
		zap.String("transponder_address", tp.ifsAddr.String()),
	)
}

func (tp *Transponder) SendProbe(probe *types.Probe, to *net.UDPAddr, finalise func(error)) {
	b, err := probe.MarshalBinary()
	if err != nil {
		finalise(err)
		return
	}

	if tp.goingDown {
		finalise(errTransponderIsAlreadyGoingDown)
		return
	}

	if tp.conn == nil {
		finalise(errTransponderIsNotConnected)
		return
	}

	go func() {
		if _, err := tp.conn.WriteToUDP(b, to); err != nil {
			// try to recover
			if err := tp.handleErrorOnSend(err); err != nil {
				finalise(err)
			}
			// try to resend
			if _, err := tp.conn.WriteToUDP(b, to); err != nil {
				// if it fails again, pass the error downstream
				finalise(err)
			}
		}
		finalise(nil)
	}()
}

func (tp *Transponder) connect() error {
	if tp.conn != nil {
		return errTransponderIsAlreadyConnected
	}
	conn, err := net.ListenUDP("udp", tp.ifsAddr)
	if err != nil {
		return err
	}
	tp.conn = conn
	return nil
}

func (tp *Transponder) reconnect() error {
	if tp.conn != nil {
		// ignore errors since we are reconnecting anyway
		_ = tp.conn.Close()
	}
	conn, err := net.ListenUDP("udp", tp.ifsAddr)
	if err != nil {
		return err
	}
	tp.conn = conn
	return nil
}

func (tp *Transponder) disconnect() error {
	if tp.conn == nil {
		return nil
	}
	return tp.conn.Close()
}

func (tp *Transponder) listen(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	if tp.conn == nil {
		failureSink <- fmt.Errorf("%s: %w",
			tp.ifsAddr, errTransponderConnectionIsNotOpen,
		)
		return
	}

	if tp.Receive == nil {
		failureSink <- fmt.Errorf("%s: %w",
			tp.ifsAddr, errTransponderReceiverIsNotSet,
		)
		return
	}

	buf := make([]byte, types.ProbeSize())

	l.Info("VPN HA-monitor transponder is going up...",
		zap.String("transponder_addr", tp.ifsAddr.String()),
	)

	for {
		length, addr, err := tp.conn.ReadFromUDP(buf)
		if err := tp.handleErrorOnReceive(err); err != nil {
			failureSink <- fmt.Errorf("%s: %w",
				tp.ifsAddr, err,
			)
			return
		}

		if length == 0 {
			return
		}

		probe := &types.Probe{}
		if err := probe.UnmarshalBinary(buf[:length]); err != nil {
			l.Warn("Failed to unmarshal incoming probe",
				zap.String("transponder_addr", tp.ifsAddr.String()),
				zap.Error(err),
			)
			continue
		}

		tp.Receive(ctx, tp, addr, probe)
	}
}

func (tp *Transponder) handleErrorOnReceive(err error) error {
	if err == nil {
		return nil
	}

	if tp.goingDown {
		return nil
	}

	if reconnectErr := tp.reconnect(); reconnectErr != nil {
		return errors.Join(err, reconnectErr)
	}

	return nil
}

func (tp *Transponder) handleErrorOnSend(err error) error {
	if err == nil {
		return nil
	}

	if tp.goingDown {
		return nil
	}

	if reconnectErr := tp.reconnect(); reconnectErr != nil {
		return errors.Join(err, reconnectErr)
	}

	return nil
}
