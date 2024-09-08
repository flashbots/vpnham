package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	errInterfaceHasNoIPs   = errors.New("interface has no ip addresses associated")
	errInterfaceHasNoIPv4s = errors.New("interface has no ipv4 addresses associated")
	errInterfaceHasNoIPv6s = errors.New("interface has no ipv4 addresses associated")
)

func GetInterfaceIPs(name string) (ipv4s []string, ipv6s []string, err error) {
	ifs, err := net.InterfaceByName(name)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w",
			name, err,
		)
	}
	addrs, err := ifs.Addrs()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w",
			name, err,
		)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet != nil {
			if ipv4 := ipNet.IP.To4(); ipv4 != nil {
				ipv4s = append(ipv4s, ipv4.String())
			} else {
				ipv6s = append(ipv6s, ipNet.IP.String())
			}
		}
	}
	if len(ipv4s) == 0 && len(ipv6s) == 0 {
		return nil, nil, fmt.Errorf("%s: %w",
			name, errInterfaceHasNoIPs,
		)
	}
	return ipv4s, ipv6s, nil
}

func GetInterfaceIP(name string, ipv4 bool) (string, error) {
	ipv4s, ipv6s, err := GetInterfaceIPs(name)
	if err != nil {
		return "", err
	}

	if ipv4 {
		if len(ipv4s) == 0 {
			return "", errInterfaceHasNoIPv4s
		}
		return ipv4s[0], nil
	} else {
		if len(ipv6s) == 0 {
			return "", errInterfaceHasNoIPv6s
		}
		return ipv6s[0], nil
	}
}

func GetInterfaceMAC(name string) (string, error) {
	ifs, err := net.InterfaceByName(name)
	if err != nil {
		return "", fmt.Errorf("%s: %w",
			name, err,
		)
	}
	return ifs.HardwareAddr.String(), nil
}

func WithTimeout(
	ctx context.Context,
	timeout time.Duration,
	do func(context.Context) error,
) error {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	err := do(ctx)
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("timed out after %v: %w", duration, err)
	}

	return err
}

func UnwrapString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func UnwrapUint32(n *uint32) uint32 {
	if n == nil {
		return 0
	}
	return *n
}

func TagsMatch(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 && len(b) != 0 {
		return false
	}
	if len(a) != 0 && len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, item := range a {
		set[item] = struct{}{}
	}
	for _, item := range b {
		if _, found := set[item]; !found {
			return false
		}
	}
	return true
}
