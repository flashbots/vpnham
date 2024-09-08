package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Client struct {
	region string

	ec2  *ec2.Client
	imds *imds.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	_imds := imds.NewFromConfig(cfg)
	regionOutput, err := _imds.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil, err
	}

	cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(regionOutput.Region))
	if err != nil {
		return nil, err
	}

	return &Client{
		region: regionOutput.Region,

		ec2:  ec2.NewFromConfig(cfg),
		imds: imds.NewFromConfig(cfg),
	}, nil
}
