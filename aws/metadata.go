package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/flashbots/vpnham/utils"
)

var (
	errFailedToDeriveRegion                = errors.New("failed to derive aws region from the environment")
	errFailedToDeriveEc2NetworkInterfaceId = errors.New("failed to derive aws ec2 instance's network interface id")
)

func Region(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveRegion, err,
		)
	}

	cli := imds.NewFromConfig(cfg)

	out, err := cli.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveRegion, err,
		)
	}

	return out.Region, nil
}

func NetworkInterfaceId(ctx context.Context, interfaceName string) (string, error) {
	ipv4s, ipv6s, err := func() (map[string]struct{}, map[string]struct{}, error) {
		_ipv4s, _ipv6s, err := utils.GetInterfaceIPs(interfaceName)
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

	macs, err := macAddresses(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveEc2NetworkInterfaceId, err,
		)
	}

	errs := []error{}
	for _, mac := range macs {
		eni, err := macInterfaceID(ctx, mac)
		if err != nil {
			return "", fmt.Errorf("%w: %w",
				errFailedToDeriveEc2NetworkInterfaceId, err,
			)
		}

		if len(ipv4s) > 0 {
			macIPv4s, err := macLocalIPv4s(ctx, mac)
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
			macIPv6s, err := macIPv6s(ctx, mac)
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

func macAddresses(ctx context.Context) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cli := imds.NewFromConfig(cfg)

	out, err := cli.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "network/interfaces/macs/",
	})
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	if _, err := io.Copy(buf, out.Content); err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(buf.String()), "\n"), nil
}

func macInterfaceID(ctx context.Context, mac string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}
	cli := imds.NewFromConfig(cfg)

	out, err := cli.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "network/interfaces/macs/" + mac + "/interface-id",
	})
	if err != nil {
		return "", err
	}

	buf := &strings.Builder{}
	if _, err := io.Copy(buf, out.Content); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func macLocalIPv4s(ctx context.Context, mac string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cli := imds.NewFromConfig(cfg)

	out, err := cli.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "network/interfaces/macs/" + mac + "/local-ipv4s",
	})
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	if _, err := io.Copy(buf, out.Content); err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(buf.String()), "\n"), nil
}

func macIPv6s(ctx context.Context, mac string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cli := imds.NewFromConfig(cfg)

	out, err := cli.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "network/interfaces/macs/" + mac + "/ipv6s",
	})
	if err != nil {
		return nil, err
	}

	buf := &strings.Builder{}
	if _, err := io.Copy(buf, out.Content); err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(buf.String()), "\n"), nil
}
