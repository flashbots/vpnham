package job

import (
	"context"
	"errors"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/utils"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
)

type updateAWSRouteTables struct {
	name    string
	timeout time.Duration

	cidr               string
	networkInterfaceID string
	routeTables        []string
}

func UpdateAWSRouteTables(
	name string,
	timeout time.Duration,
	cidr string,
	networkInterfaceID string,
	routeTables []string,
) Job {
	return &updateAWSRouteTables{
		name:               name,
		timeout:            timeout,
		cidr:               cidr,
		networkInterfaceID: networkInterfaceID,
		routeTables:        routeTables,
	}
}

func (j *updateAWSRouteTables) Name() string {
	return j.name
}

func (j *updateAWSRouteTables) Execute(ctx context.Context) error {
	errs := []error{}
	for _, rt := range j.routeTables {
		err := j.updateRouteTable(ctx, rt, j.cidr, j.networkInterfaceID)
		if err != nil {
			metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
				attribute.String(metrics.LabelErrorScope, "job_"+j.name),
			))
			errs = append(errs, err)
		}
	}

	switch len(errs) {
	default:
		return errors.Join(errs...)
	case 1:
		return errs[0]
	case 0:
		return nil
	}
}

func (j *updateAWSRouteTables) updateRouteTable(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	cli, err := aws.NewClient(ctx)
	if err != nil {
		return err
	}

	var route *awstypes.Route

	// check if the route is already set
	err = utils.WithTimeout(ctx, j.timeout, func(ctx context.Context) error {
		route, err = cli.FindRoute(ctx, routeTable, cidr)
		return err
	})
	if err != nil {
		return err
	}

	if route != nil && awssdk.ToString(route.NetworkInterfaceId) == networkInterfaceID {
		// route is already up to date
		return nil
	}

	if route != nil {
		// route exists but with different next hop
		return utils.WithTimeout(ctx, j.timeout, func(ctx context.Context) error {
			return cli.UpdateRoute(ctx, routeTable, cidr, networkInterfaceID)
		})
	}

	// no route yet
	return utils.WithTimeout(ctx, j.timeout, func(ctx context.Context) error {
		return cli.CreateRoute(ctx, routeTable, cidr, networkInterfaceID)
	})
}
