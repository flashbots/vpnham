package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	awscli "github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/utils"
)

type ReconcileBridgeActivateAWS struct {
	BridgeInterface     string                                    `yaml:"-"`
	Region              string                                    `yaml:"-"`
	SecondaryInterfaces []string                                  `yaml:"-"`
	Vpcs                map[string]*ReconcileBridgeActivateAWSVpc `yaml:"-"`

	Timeout time.Duration `yaml:"timeout"`

	RouteTables []string `yaml:"route_tables"`
}

type ReconcileBridgeActivateAWSVpc struct {
	ID                 string
	LocalInterfaceID   string
	NetworkInterfaceID string
	RouteTables        []string
}

var (
	errAWSDuplicateVpcForSecondaryInterface = errors.New("secondary interface belongs to a vpc that another interface is already attached")
	errAWSRouteTableWithoutInterface        = errors.New("route table has not interface attached to its vpc")
)

func (r *ReconcileBridgeActivateAWS) PostLoad(ctx context.Context) error {
	if r.Timeout == 0 {
		r.Timeout = DefaultAWSTimeout
	}

	{ // aws region
		var reg string
		err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			reg, err = awscli.Region(ctx)
			return err
		})
		if err != nil {
			return err
		}
		r.Region = reg
	}

	aws, err := awscli.NewClient(ctx)
	if err != nil {
		return err
	}

	{ // aws ec2 network interface id
		var networkInterfaceID, vpcID string
		err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			networkInterfaceID, err = aws.NetworkInterfaceId(ctx, r.BridgeInterface)
			return err
		})
		if err != nil {
			return err
		}
		err = utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			vpcID, err = aws.NetworkInterfaceVpcID(ctx, networkInterfaceID)
			return err
		})
		if err != nil {
			return err
		}
		r.Vpcs = make(map[string]*ReconcileBridgeActivateAWSVpc, 1+len(r.SecondaryInterfaces))
		r.Vpcs[vpcID] = &ReconcileBridgeActivateAWSVpc{
			ID:                 vpcID,
			LocalInterfaceID:   r.BridgeInterface,
			NetworkInterfaceID: networkInterfaceID,
			RouteTables:        make([]string, 0),
		}
	}

	{ // aws ec2 secondary interface ids
		for _, ifs := range r.SecondaryInterfaces {
			var networkInterfaceID, vpcID string
			err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
				networkInterfaceID, err = aws.NetworkInterfaceId(ctx, ifs)
				return err
			})
			if err != nil {
				return err
			}
			err = utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
				vpcID, err = aws.NetworkInterfaceVpcID(ctx, networkInterfaceID)
				return err
			})
			if err != nil {
				return err
			}
			if dupe, exists := r.Vpcs[vpcID]; exists {
				return fmt.Errorf("%w: vpc %s, interface %s, duplicate %s",
					errAWSDuplicateVpcForSecondaryInterface, vpcID, dupe.LocalInterfaceID, ifs,
				)
			}
			r.Vpcs[vpcID] = &ReconcileBridgeActivateAWSVpc{
				ID:                 vpcID,
				LocalInterfaceID:   ifs,
				NetworkInterfaceID: networkInterfaceID,
				RouteTables:        make([]string, 0),
			}
		}
	}

	{ // route tables
		for _, routeTable := range r.RouteTables {
			var vpcID string
			err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
				vpcID, err = aws.RouteTableVpcID(ctx, routeTable)
				return err
			})
			if err != nil {
				return err
			}
			vpc, exists := r.Vpcs[vpcID]
			if !exists {
				return fmt.Errorf("%w: %s",
					errAWSRouteTableWithoutInterface, routeTable,
				)
			}
			vpc.RouteTables = append(vpc.RouteTables, routeTable)
		}
	}

	return nil
}
