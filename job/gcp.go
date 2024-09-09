package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	gcepb "cloud.google.com/go/compute/apiv1/computepb"
	gcpcli "github.com/flashbots/vpnham/gcp"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"
)

type UpdateGCPRoute struct {
	JobName string
	Timeout time.Duration

	Name        string
	Description string

	DestRanges      []types.CIDR
	Network         string
	NextHopInstance string
	Priority        uint32
	Tags            []string
}

func (j *UpdateGCPRoute) GetJobName() string {
	return j.JobName
}

func (j *UpdateGCPRoute) Execute(ctx context.Context) error {
	errs := make([]error, 0)
	for idx, destRange := range j.DestRanges {
		if err := j.updateRoute(ctx, idx, destRange.String()); err != nil {
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

func (j *UpdateGCPRoute) updateRoute(
	ctx context.Context,
	idx int,
	destRange string,
) error {
	gcp, err := gcpcli.NewClient(ctx)
	if err != nil {
		return err
	}

	var routes []*gcepb.Route
	err = utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) (err error) {
		routes, err = gcp.FindRoute(ctx, j.Network, destRange)
		return err
	})
	if err != nil {
		metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
			attribute.String(metrics.LabelErrorScope, "job_"+j.JobName),
		))
		return err
	}

	switch len(routes) {
	case 0:
		// no route yet
		return utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return gcp.CreateRoute(ctx, j.gceRoute(idx, destRange))
		})

	case 1:
		route := routes[0]
		if j.matches(idx, route) {
			// route is already up to date
			return nil
		}
		// route exists but with different config => delete, then create
		err := utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return gcp.DeleteRoute(ctx, route)
		})
		if err != nil {
			return err
		}
		return utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
			return gcp.CreateRoute(ctx, j.gceRoute(idx, destRange))
		})

	default:
		// delete all non-matching routes
		errs := make([]error, 0)
		foundMatch := false
		for _, route := range routes {
			if foundMatch {
				// we already found matching rule, so let's clean up the rest
				err := utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
					return gcp.DeleteRoute(ctx, route)
				})
				if err != nil {
					errs = append(errs, err)
				}
				continue
			}
			if j.matches(idx, route) {
				foundMatch = true
				continue
			}
			err := utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
				return gcp.DeleteRoute(ctx, route)
			})
			if err != nil {
				errs = append(errs, err)
			}
		}

		// if the match not found, create a new one
		if !foundMatch {
			err := utils.WithTimeout(ctx, j.Timeout, func(ctx context.Context) error {
				return gcp.CreateRoute(ctx, j.gceRoute(idx, destRange))
			})
			if err != nil {
				errs = append(errs, err)
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
}

func (j *UpdateGCPRoute) gceRoute(idx int, destRange string) *gcepb.Route {
	name := j.Name + "-" + fmt.Sprintf("%02d", idx)

	return &gcepb.Route{
		Name:        proto.String(name),
		Description: proto.String(j.Description),

		DestRange:       proto.String(destRange),
		Network:         proto.String(j.Network),
		NextHopInstance: proto.String(j.NextHopInstance),
		Priority:        proto.Uint32(j.Priority),
		Tags:            j.Tags,
	}
}

func (j *UpdateGCPRoute) matches(idx int, route *gcepb.Route) bool {
	return utils.UnwrapString(route.Name) == j.Name+"-"+fmt.Sprintf("%02d", idx) &&
		utils.UnwrapString(route.NextHopInstance) == j.NextHopInstance &&
		utils.UnwrapUint32(route.Priority) == j.Priority &&
		utils.TagsMatch(route.Tags, j.Tags)
}
