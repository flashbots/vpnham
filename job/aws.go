package job

import (
	"context"
	"errors"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awscli "github.com/flashbots/vpnham/aws"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
)

type UpdateAWSRouteTables struct {
	aws *awscli.Client

	JobName string
	Timeout time.Duration

	DestinationCidrBlocks []types.CIDR
	NetworkInterfaceID    string
	RouteTables           []string
}

func (j *UpdateAWSRouteTables) GetJobName() string {
	return j.JobName
}

func (j *UpdateAWSRouteTables) Execute(ctx context.Context) error {
	aws, err := awscli.NewClient(ctx)
	if err != nil {
		return err
	}
	j.aws = aws

	errs := []error{}
	for _, destinationCidrBlock := range j.DestinationCidrBlocks {
		for _, rt := range j.RouteTables {
			err := j.updateRouteTable(ctx, rt, destinationCidrBlock.String(), j.NetworkInterfaceID)
			if err != nil {
				metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
					attribute.String(metrics.LabelErrorScope, "job_"+j.JobName),
				))
				errs = append(errs, err)
			}
		}
	}

	switch len(errs) {
	default:
		metrics.Errors.Add(ctx, int64(len(errs)), otelapi.WithAttributes(
			attribute.String(metrics.LabelErrorScope, "job_"+j.JobName),
		))
		return errors.Join(errs...)
	case 1:
		metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelErrorScope, "job_"+j.JobName),
		))
		return errs[0]
	case 0:
		return nil
	}
}

func (j *UpdateAWSRouteTables) updateRouteTable(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	var routes []*awstypes.Route
	err := utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) (err error) {
		routes, err = j.aws.FindRoute(ctx, routeTable, cidr)
		return err
	})
	if err != nil {
		return err
	}

	switch len(routes) {
	case 0:
		// no route yet
		return utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return j.aws.CreateRoute(ctx, routeTable, cidr, networkInterfaceID)
		})

	case 1:
		route := routes[0]
		if awssdk.ToString(route.NetworkInterfaceId) == networkInterfaceID {
			// route is already up to date
			return nil
		}
		// route exists but with different next hop
		return utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return j.aws.UpdateRoute(ctx, routeTable, route, cidr, networkInterfaceID)
		})

	default:
		// i.d.k. if this is even possible to have 2+ routes with the same
		// destination cidr in aws route-table.  but at any rate, if that's the
		// case let's just delete all of them and create a new one
		for _, route := range routes {
			err = utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
				return j.aws.DeleteRoute(ctx, routeTable, route)
			})
			if err != nil {
				return err
			}
		}
		return utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return j.aws.CreateRoute(ctx, routeTable, cidr, networkInterfaceID)
		})
	}
}
