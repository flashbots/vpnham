package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

var (
	errFailedToDeriveRegion = errors.New("failed to derive aws region from the environment")
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
