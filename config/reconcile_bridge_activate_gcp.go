package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	gcpcli "github.com/flashbots/vpnham/gcp"
	"github.com/flashbots/vpnham/utils"
)

type ReconcileBridgeActivateGCP struct {
	BridgeName          string                                    `yaml:"-"`
	BridgeInterface     string                                    `yaml:"-"`
	InstanceName        string                                    `yaml:"-"`
	ProjectID           string                                    `yaml:"-"`
	SecondaryInterfaces []string                                  `yaml:"-"`
	Vpcs                map[string]*ReconcileBridgeActivateGCPVpc `yaml:"-"`

	RouteIDPrefix string   `yaml:"route_id_prefix"`
	RoutePriority uint32   `yaml:"route_priority"`
	RouteTags     []string `yaml:"route_tags"`

	Timeout time.Duration `yaml:"timeout"`
}

type ReconcileBridgeActivateGCPVpc struct {
	ID               string
	LocalInterfaceID string
}

var (
	errGCPDuplicateVpcForSecondaryInterface = errors.New("secondary interface belongs to a vpc that another interface is already attached to")
)

func (r *ReconcileBridgeActivateGCP) PostLoad(ctx context.Context) error {
	if r.Timeout == 0 {
		r.Timeout = DefaultGCPTimeout
	}

	if r.RouteIDPrefix == "" {
		r.RouteIDPrefix = DefaultRouteIDPrefix + "-" + r.BridgeName
	}

	if r.RoutePriority == 0 {
		r.RoutePriority = DefaultGCPRoutePriority
	}

	gcp, err := gcpcli.NewClient(ctx)
	if err != nil {
		return err
	}

	{ // project id
		var projectID string
		err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			projectID, err = gcp.ProjectID(ctx)
			return err
		})
		if err != nil {
			return err
		}
		r.ProjectID = projectID
	}

	{ // instance name
		var instanceName string
		err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			instanceName, err = gcp.InstanceName(ctx)
			return err
		})
		if err != nil {
			return err
		}
		r.InstanceName = gcp.NormaliseInstanceName(instanceName)
	}

	{ // gcp gce networks
		var vpcID string
		err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
			vpcID, err = gcp.NetworkInterfaceVpcID(ctx, r.BridgeInterface)
			return err
		})
		if err != nil {
			return err
		}
		r.Vpcs = make(map[string]*ReconcileBridgeActivateGCPVpc, 1+len(r.SecondaryInterfaces))
		r.Vpcs[vpcID] = &ReconcileBridgeActivateGCPVpc{
			ID:               gcp.NormaliseNetworkID(vpcID),
			LocalInterfaceID: r.BridgeInterface,
		}
	}

	{ // gcp secondary gce networks
		for _, ifs := range r.SecondaryInterfaces {
			var vpcID string
			err := utils.WithTimeout(ctx, r.Timeout, func(ctx context.Context) (err error) {
				vpcID, err = gcp.NetworkInterfaceVpcID(ctx, ifs)
				return err
			})
			if err != nil {
				return err
			}
			if dupe, exists := r.Vpcs[vpcID]; exists {
				return fmt.Errorf("%w: vpc %s, interface %s, duplicate %s",
					errGCPDuplicateVpcForSecondaryInterface, vpcID, dupe.LocalInterfaceID, ifs,
				)
			}
			r.Vpcs[vpcID] = &ReconcileBridgeActivateGCPVpc{
				ID:               vpcID,
				LocalInterfaceID: ifs,
			}
		}
	}

	return nil
}
