package types

import (
	"errors"
	"fmt"
	"net"
)

type CIDR string

var (
	errCIDRIsInvalid = errors.New("cidr is invalid")
)

func (c CIDR) String() string {
	return string(c)
}

func (c CIDR) Validate() error {
	if _, _, err := net.ParseCIDR(string(c)); err != nil {
		return fmt.Errorf("%w: %w",
			errCIDRIsInvalid, err,
		)
	}
	return nil
}

func (c CIDR) IsIPv4() bool {
	addr, _, err := net.ParseCIDR(string(c))
	if err != nil {
		return false
	}
	if addr.To4() != nil {
		return true
	}
	return false
}
