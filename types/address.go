package types

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Address string

var (
	errAddressInvalidIP   = errors.New("ip is invalid")
	errAddressInvalidPort = errors.New("port is invalid")
	errAddressMissingPort = errors.New("missing port")
)

func (a Address) Parse() (ip net.IP, port int, err error) {
	parts := strings.Split(string(a), ":")
	if len(parts) == 1 {
		return nil, 0, fmt.Errorf("%w: %s",
			errAddressMissingPort, a,
		)
	}

	strPort := parts[len(parts)-1]
	strIP := strings.Join(parts[:len(parts)-1], ":")

	port, err = strconv.Atoi(strPort)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w",
			errAddressInvalidPort, err,
		)
	}
	if port <= 0 || port > 65535 {
		return nil, 0, fmt.Errorf("%w: %s: %d",
			errAddressInvalidPort, "out of range", port,
		)
	}

	strIP = strings.TrimPrefix(strings.TrimSuffix(strIP, "]"), "[")
	ip = net.ParseIP(strIP)
	if ip == nil {
		return nil, 0, fmt.Errorf("%w: %s",
			errAddressInvalidIP, strIP,
		)
	}

	return ip, port, nil
}

func (a Address) Validate() error {
	if _, _, err := a.Parse(); err != nil {
		return err
	}
	return nil
}
