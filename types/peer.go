package types

import (
	"net"

	"github.com/google/uuid"
)

type Peer struct {
	iname string

	probeAddr *net.UDPAddr
	uuid      uuid.UUID

	sequence        uint64
	acknowledgement uint64
}

func NewPeer(iname string, addr Address) (*Peer, error) {
	ip, port, err := addr.Parse()
	if err != nil {
		return nil, err
	}
	probeAddr := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	_uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Peer{
		iname: iname,

		probeAddr: probeAddr,
		uuid:      _uuid,
	}, nil
}

func (p *Peer) InterfaceName() string {
	return p.iname
}

func (p *Peer) Sequence() uint64 {
	return p.sequence
}

func (p *Peer) NextSequence() uint64 {
	p.sequence += 1
	return p.sequence
}

func (p *Peer) SetAcknowledgement(ack uint64) {
	p.acknowledgement = ack
}

func (p *Peer) Acknowledgement() uint64 {
	return p.acknowledgement
}

func (p *Peer) ProbeAddr() *net.UDPAddr {
	return p.probeAddr
}

func (p *Peer) UUID() uuid.UUID {
	return p.uuid
}
