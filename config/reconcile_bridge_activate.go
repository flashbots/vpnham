package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
)

type ReconcileBridgeActivate struct {
	BridgeInterface string `yaml:"-"`

	Reapply *ReconcileReapply `yaml:"reapply"`

	AWS    *ReconcileBridgeActivateAWS `yaml:"aws"`
	Script types.Script                `yaml:"script"`
}

type ReconcileBridgeActivateAWS struct {
	NetworkInterfaceID string `yaml:"-"`
	Region             string `yaml:"-"`

	Timeout time.Duration `yaml:"timeout"`

	RouteTables []string `yaml:"route_tables"`
}

var (
	errAWSRouteTableDoesNotExist = errors.New("aws route table does not exist")
)

func (r *ReconcileBridgeActivate) PostLoad(ctx context.Context) error {
	if r.Reapply == nil {
		r.Reapply = &ReconcileReapply{}
	}

	if err := r.Reapply.PostLoad(ctx); err != nil {
		return err
	}

	if r.AWS != nil {
		if r.AWS.Timeout == 0 {
			r.AWS.Timeout = DefaultAWSTimeout
		}

		{ // aws region
			var reg string
			err := utils.WithTimeout(ctx, r.AWS.Timeout, func(ctx context.Context) (err error) {
				reg, err = aws.Region(ctx)
				return err
			})
			if err != nil {
				return err
			}
			r.AWS.Region = reg
		}

		{ // aws ec2 network interface id
			var networkInterfaceID string
			err := utils.WithTimeout(ctx, r.AWS.Timeout, func(ctx context.Context) (err error) {
				networkInterfaceID, err = aws.NetworkInterfaceId(ctx, r.BridgeInterface)
				return err
			})
			if err != nil {
				return err
			}
			r.AWS.NetworkInterfaceID = networkInterfaceID
		}
	}

	return nil
}

func (r *ReconcileBridgeActivate) Validate(ctx context.Context) error {
	if err := r.Reapply.Validate(ctx); err != nil {
		return err
	}

	if r.AWS != nil {
		cli, err := aws.NewClient(ctx)
		if err != nil {
			return err
		}

		for _, rt := range r.AWS.RouteTables {
			exists, err := cli.RouteTableExists(ctx, rt)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("%w: %s",
					errAWSRouteTableDoesNotExist, rt,
				)
			}
		}
	}

	return nil
}
