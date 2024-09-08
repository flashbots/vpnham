package gcp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/flashbots/vpnham/utils"
)

var (
	errFailedToDeriveVpcIdFromInterfaceName = errors.New("failed to derive network id from local interface name")
)

func (cli *Client) NetworkInterfaceVpcID(
	ctx context.Context,
	localInterfaceName string,
) (string, error) {
	if !metadata.OnGCE() {
		return "", errGCPNotOnGCE
	}

	mac, err := utils.GetInterfaceMAC(localInterfaceName)
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromInterfaceName, err,
		)
	}

	networkInterfaces, err := metadata.GetWithContext(ctx, "instance/network-interfaces/")
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromInterfaceName, err,
		)
	}

	for _, ifs := range strings.Split(networkInterfaces, "\n") {
		ifs = strings.TrimSuffix(ifs, "/")
		if ifs == "" {
			continue
		}

		ifsMac, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/network-interfaces/%s/mac", ifs))
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveVpcIdFromInterfaceName, err,
			)
		}

		if ifsMac != mac {
			continue
		}

		vpcID, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/network-interfaces/%s/network", ifs))
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveVpcIdFromInterfaceName, err,
			)
		}

		return vpcID, nil
	}

	return "", fmt.Errorf("%w: interface has no gce network: %s",
		errFailedToDeriveVpcIdFromInterfaceName, localInterfaceName,
	)
}

func (cli *Client) NetworkInterfaceVpcIP(
	ctx context.Context,
	localInterfaceName string,
) (string, error) {
	mac, err := utils.GetInterfaceMAC(localInterfaceName)
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromInterfaceName, err,
		)
	}

	networkInterfaces, err := metadata.GetWithContext(ctx, "instance/network-interfaces/")
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromInterfaceName, err,
		)
	}

	for _, ifs := range strings.Split(networkInterfaces, "\n") {
		ifs = strings.TrimSuffix(ifs, "/")
		if ifs == "" {
			continue
		}

		ifsMac, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/network-interfaces/%s/mac", ifs))
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveVpcIdFromInterfaceName, err,
			)
		}

		if ifsMac != mac {
			continue
		}

		ip, err := metadata.GetWithContext(ctx, fmt.Sprintf("instance/network-interfaces/%s/ip", ifs))
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveVpcIdFromInterfaceName, err,
			)
		}

		return ip, nil
	}

	return "", fmt.Errorf("%w: interface has no gce network: %s",
		errFailedToDeriveVpcIdFromInterfaceName, localInterfaceName,
	)
}
