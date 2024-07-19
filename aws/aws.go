package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var (
	errRouteTableDoesNotExist = errors.New("aws route table does not exist")
)

type Client struct {
	ec2 *ec2.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	reg, err := func() (string, error) {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", err
		}

		cli := imds.NewFromConfig(cfg)

		out, err := cli.GetRegion(ctx, &imds.GetRegionInput{})
		if err != nil {
			return "", err
		}

		return out.Region, nil
	}()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(reg))
	if err != nil {
		return nil, err
	}

	return &Client{
		ec2: ec2.NewFromConfig(cfg),
	}, nil
}
