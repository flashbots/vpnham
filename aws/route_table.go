package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/flashbots/vpnham/logutils"
	"go.uber.org/zap"
)

func (cli *Client) RouteTableExists(ctx context.Context, routeTable string) (bool, error) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Describing AWS route table...",
		zap.String("route_table", routeTable),
	)

	out, err := cli.ec2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTable},
	})
	if err != nil {
		l.Error("Failed to describe AWS route table",
			zap.Error(err),
			zap.String("route_table", routeTable),
		)
		return false, err
	}

	if len(out.RouteTables) == 0 {
		return false, nil
	}

	return true, nil
}

func (cli *Client) UpdateRouteTable(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	// check if the route is already set
	route, err := cli.findRoute(ctx, routeTable, cidr)
	if err != nil {
		return err
	}

	if route != nil && aws.ToString(route.NetworkInterfaceId) == networkInterfaceID {
		// route is already up to date
		return nil
	}

	if route != nil {
		// route exists but with different next hop
		return cli.updateRoute(ctx, routeTable, cidr, networkInterfaceID)
	}

	// no route yet
	return cli.createRoute(ctx, routeTable, cidr, networkInterfaceID)
}

func (cli *Client) findRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
) (*awstypes.Route, error) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Describing AWS route table...",
		zap.String("route_table", routeTable),
	)

	out, err := cli.ec2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTable},
	})
	if err != nil {
		l.Error("Failed to describe AWS route table",
			zap.Error(err),
			zap.String("route_table", routeTable),
		)
		return nil, err
	}

	if len(out.RouteTables) == 0 {
		return nil, fmt.Errorf("%w: %s",
			errRouteTableDoesNotExist, routeTable,
		)
	}
	rt := out.RouteTables[0]

	for _, route := range rt.Routes {
		if aws.ToString(route.DestinationCidrBlock) == cidr {
			return &route, nil
		}
	}

	return nil, nil
}

func (cli *Client) updateRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Replacing route in AWS route table...",
		zap.String("cidr", cidr),
		zap.String("network_interface_id", networkInterfaceID),
		zap.String("route_table", routeTable),
	)

	_, err := cli.ec2.ReplaceRoute(ctx, &ec2.ReplaceRouteInput{
		RouteTableId:         aws.String(routeTable),
		DestinationCidrBlock: aws.String(cidr),
		NetworkInterfaceId:   aws.String(networkInterfaceID),
	})
	if err != nil {
		l.Error("Failed to replace route in AWS route table",
			zap.Error(err),
			zap.String("cidr", cidr),
			zap.String("network_interface_id", networkInterfaceID),
			zap.String("route_table", routeTable),
		)
	}
	return err
}

func (cli *Client) createRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Creating route in AWS route table...",
		zap.String("cidr", cidr),
		zap.String("network_interface_id", networkInterfaceID),
		zap.String("route_table", routeTable),
	)

	_, err := cli.ec2.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         aws.String(routeTable),
		DestinationCidrBlock: aws.String(cidr),
		NetworkInterfaceId:   aws.String(networkInterfaceID),
	})
	if err != nil {
		l.Error("Failed to create route in AWS route table",
			zap.Error(err),
			zap.String("cidr", cidr),
			zap.String("network_interface_id", networkInterfaceID),
			zap.String("route_table", routeTable),
		)
	}
	return err
}
