package aws

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

func (cli *Client) macAddresses(ctx context.Context) ([]string, error) {
	out, err := cli.imds.GetMetadata(ctx, &imds.GetMetadataInput{
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

func (cli *Client) macInterfaceID(ctx context.Context, mac string) (string, error) {
	out, err := cli.imds.GetMetadata(ctx, &imds.GetMetadataInput{
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

func (cli *Client) macLocalIPv4s(ctx context.Context, mac string) ([]string, error) {
	out, err := cli.imds.GetMetadata(ctx, &imds.GetMetadataInput{
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

func (cli *Client) macIPv6s(ctx context.Context, mac string) ([]string, error) {
	out, err := cli.imds.GetMetadata(ctx, &imds.GetMetadataInput{
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
