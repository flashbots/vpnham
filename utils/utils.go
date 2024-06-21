package utils

import (
	"errors"
	"fmt"
	"net"
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
