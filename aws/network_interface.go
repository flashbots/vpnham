package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/flashbots/vpnham/utils"
)

var (
	errFailedToDeriveEc2NetworkInterfaceId = errors.New("failed to derive aws ec2 instance's network interface id")
	errFailedToDeriveVpcIdFromInterface    = errors.New("failed to derive vpc id from network interface id")
)

func (cli *Client) NetworkInterfaceId(
	ctx context.Context,
	localInterfaceName string,
) (string, error) {
	ipv4s, ipv6s, err := func() (map[string]struct{}, map[string]struct{}, error) {
		_ipv4s, _ipv6s, err := utils.GetInterfaceIPs(localInterfaceName)
		if err != nil {
			return nil, nil, err
		}

		ipv4s := make(map[string]struct{}, len(_ipv4s))
		ipv6s := make(map[string]struct{}, len(_ipv6s))

		for _, ipv4 := range _ipv4s {
			ipv4s[ipv4] = struct{}{}
		}

		for _, ipv6 := range _ipv6s {
			ipv6s[ipv6] = struct{}{}
		}

		return ipv4s, ipv6s, nil
	}()
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveEc2NetworkInterfaceId, err,
		)
	}

	macs, err := cli.macAddresses(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveEc2NetworkInterfaceId, err,
		)
	}

	errs := []error{}
	for _, mac := range macs {
		eni, err := cli.macInterfaceID(ctx, mac)
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveEc2NetworkInterfaceId, err,
			)
		}

		if len(ipv4s) > 0 {
			macIPv4s, err := cli.macLocalIPv4s(ctx, mac)
			if err != nil {
				// there could be legit errors
				// (e.g. when getting ipv4 addresses for ipv6-only interface)
				errs = append(errs, err)
			} else {
				for _, ipv4 := range macIPv4s {
					if _, exists := ipv4s[ipv4]; exists {
						return eni, nil
					}
				}
			}
		}

		if len(ipv6s) > 0 {
			macIPv6s, err := cli.macIPv6s(ctx, mac)
			if err != nil {
				// there could be legit errors
				// (e.g. when getting ipv6 addresses for ipv4-only interface)
				errs = append(errs, err)
			} else {
				for _, ipv6 := range macIPv6s {
					if _, exists := ipv6s[ipv6]; exists {
						return eni, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("%w: %w",
		errFailedToDeriveEc2NetworkInterfaceId,
		fmt.Errorf("could not find matching ip: %w", errors.Join(errs...)),
	)
}

func (cli *Client) NetworkInterfaceVpcID(
	ctx context.Context,
	interfaceID string,
) (string, error) {
	out, err := cli.ec2.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{interfaceID},
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromInterface, err,
		)
	}

	if len(out.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("%w: interface not found: %s",
			errFailedToDeriveVpcIdFromInterface, interfaceID,
		)
	}
	ifs := out.NetworkInterfaces[0]

	if ifs.VpcId == nil {
		return "", fmt.Errorf("%w: interface has no vpc id: %s",
			errFailedToDeriveVpcIdFromInterface, interfaceID,
		)
	}

	return *ifs.VpcId, nil
}
