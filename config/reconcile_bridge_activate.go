package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/types"
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
		reg, err := aws.Region(ctx)
		if err != nil {
			return err
		}
		r.AWS.Region = reg

		eni, err := aws.NetworkInterfaceId(ctx, r.BridgeInterface)
		if err != nil {
			return err
		}
		r.AWS.NetworkInterfaceID = eni
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
