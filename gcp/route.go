package gcp

import (
	"context"
	"fmt"

	gcepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/utils"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

func (cli *Client) FindRoute(
	ctx context.Context,
	networkID string,
	cidr string,
) ([]*gcepb.Route, error) {
	l := logutils.LoggerFromContext(ctx)

	filter := fmt.Sprintf(`network = "%s" AND destRange = "%s"`,
		networkID, cidr,
	)

	l.Debug("Listing GCP routes...",
		zap.String("filter", filter),
		zap.String("project", cli.projectID),
	)

	iter := cli.routes.List(ctx, &gcepb.ListRoutesRequest{
		Filter:  proto.String(filter),
		Project: cli.projectID,
	})

	routes := make([]*gcepb.Route, 0, 1)
	for {
		route, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			l.Error("Failed to list GCP routes",
				zap.Error(err),
				zap.String("filter", filter),
				zap.String("project", cli.projectID),
			)
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

func (cli *Client) CreateRoute(
	ctx context.Context,
	route *gcepb.Route,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Info("Inserting GCP route...",
		zap.String("description", utils.UnwrapString(route.Description)),
		zap.String("dest_range", utils.UnwrapString(route.DestRange)),
		zap.String("name", utils.UnwrapString(route.Name)),
		zap.String("network", utils.UnwrapString(route.Network)),
		zap.String("next_hop_instance", utils.UnwrapString(route.NextHopInstance)),
		zap.String("project", cli.projectID),
		zap.Strings("tags", route.Tags),
		zap.Uint32("priority", utils.UnwrapUint32(route.Priority)),
	)

	err := func() error {
		op, err := cli.routes.Insert(ctx, &gcepb.InsertRouteRequest{
			Project:       cli.projectID,
			RouteResource: route,
		})
		if err != nil {
			return err
		}
		return op.Wait(ctx)
	}()
	if err != nil {
		l.Error("Failed to insert GCP route",
			zap.Error(err),
			zap.String("description", utils.UnwrapString(route.Description)),
			zap.String("dest_range", utils.UnwrapString(route.DestRange)),
			zap.String("name", utils.UnwrapString(route.Name)),
			zap.String("network", utils.UnwrapString(route.Network)),
			zap.String("next_hop_instance", utils.UnwrapString(route.NextHopInstance)),
			zap.String("project", cli.projectID),
			zap.Strings("tags", route.Tags),
			zap.Uint32("priority", utils.UnwrapUint32(route.Priority)),
		)
	}
	return err
}

func (cli *Client) DeleteRoute(
	ctx context.Context,
	route *gcepb.Route,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Warn("Deleting GCP route...",
		// project
		zap.String("project", cli.projectID),
		// network
		zap.String("network", utils.UnwrapString(route.Network)),
		// route
		zap.String("name", utils.UnwrapString(route.Name)),
		zap.String("kind", utils.UnwrapString(route.Kind)),
		zap.String("route_type", utils.UnwrapString(route.RouteType)),
		// destination
		zap.String("dest_range", utils.UnwrapString(route.DestRange)),
		// next hop
		zap.String("next_hop_gateway", utils.UnwrapString(route.NextHopGateway)),
		zap.String("next_hop_hub", utils.UnwrapString(route.NextHopHub)),
		zap.String("next_hop_ilb", utils.UnwrapString(route.NextHopIlb)),
		zap.String("next_hop_instance", utils.UnwrapString(route.NextHopInstance)),
		zap.String("next_hop_ip", utils.UnwrapString(route.NextHopIp)),
		zap.String("next_hop_network", utils.UnwrapString(route.NextHopNetwork)),
		zap.String("next_hop_peering", utils.UnwrapString(route.NextHopPeering)),
		zap.String("next_hop_vpn_tunnel", utils.UnwrapString(route.NextHopVpnTunnel)),
		// priority
		zap.Uint32("priority", utils.UnwrapUint32(route.Priority)),
		// tags
		zap.Strings("tags", route.Tags),
		// rest
		zap.String("route_status", utils.UnwrapString(route.RouteStatus)),
	)

	err := func() error {
		op, err := cli.routes.Delete(ctx, &gcepb.DeleteRouteRequest{
			Project: cli.projectID,
			Route:   utils.UnwrapString(route.Name),
		})
		if err != nil {
			return err
		}
		return op.Wait(ctx)
	}()
	if err != nil {
		l.Error("Failed to delete GCP route...",
			zap.Error(err),
			// project
			zap.String("project", cli.projectID),
			// network
			zap.String("network", utils.UnwrapString(route.Network)),
			// route
			zap.String("name", utils.UnwrapString(route.Name)),
			zap.String("kind", utils.UnwrapString(route.Kind)),
			zap.String("route_type", utils.UnwrapString(route.RouteType)),
			// destination
			zap.String("dest_range", utils.UnwrapString(route.DestRange)),
			// next hop
			zap.String("next_hop_gateway", utils.UnwrapString(route.NextHopGateway)),
			zap.String("next_hop_hub", utils.UnwrapString(route.NextHopHub)),
			zap.String("next_hop_ilb", utils.UnwrapString(route.NextHopIlb)),
			zap.String("next_hop_instance", utils.UnwrapString(route.NextHopInstance)),
			zap.String("next_hop_ip", utils.UnwrapString(route.NextHopIp)),
			zap.String("next_hop_network", utils.UnwrapString(route.NextHopNetwork)),
			zap.String("next_hop_peering", utils.UnwrapString(route.NextHopPeering)),
			zap.String("next_hop_vpn_tunnel", utils.UnwrapString(route.NextHopVpnTunnel)),
			// priority
			zap.Uint32("priority", utils.UnwrapUint32(route.Priority)),
			// tags
			zap.Strings("tags", route.Tags),
			// rest
			zap.String("route_status", utils.UnwrapString(route.RouteStatus)),
		)
	}
	return err
}
