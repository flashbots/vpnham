package types

import (
	"errors"
	"fmt"
	"net/url"
)

type Partner struct {
	url *url.URL

	sequence uint64
}

var (
	errPartnerURLIsInvalid = errors.New("bridge partner url is invalid")
)

func NewPartner(partnerURL string) (*Partner, error) {
	_url, err := url.Parse(partnerURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errPartnerURLIsInvalid, err,
		)
	}

	return &Partner{
		url: _url,
	}, nil
}

func (p *Partner) URL() *url.URL {
	return p.url
}

func (p *Partner) Sequence() uint64 {
	return p.sequence
}

func (p *Partner) NextSequence() uint64 {
	p.sequence += 1
	return p.sequence
}
