package types

import (
	"errors"
	"fmt"
	"net/url"
)

type Partner struct {
	statusURL *url.URL

	sequence uint64
}

var (
	errPartnerStatusURLIsInvalid = errors.New("bridge partner status url is invalid")
)

func NewPartner(statusURL string) (*Partner, error) {
	_statusURL, err := url.Parse(statusURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errPartnerStatusURLIsInvalid, err,
		)
	}

	return &Partner{
		statusURL: _statusURL,
	}, nil
}

func (p *Partner) StatusURL() *url.URL {
	return p.statusURL
}

func (p *Partner) Sequence() uint64 {
	return p.sequence
}

func (p *Partner) NextSequence() uint64 {
	p.sequence += 1
	return p.sequence
}
