package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/flashbots/vpnham/logutils"
	"go.uber.org/zap"
)

var (
	errFailedToDeriveVpcIdFromRouteTable = errors.New("failed to derive vpc id from route-table id")
	errRouteTableDoesNotExist            = errors.New("aws route-table does not exist")
)

func (cli *Client) RouteTableVpcID(
	ctx context.Context,
	routeTable string,
) (string, error) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Describing AWS route-table...",
		zap.String("route_table_id", routeTable),
	)

	out, err := cli.ec2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTable},
	})
	if err != nil {
		l.Error("Failed to describe AWS route-table",
			zap.Error(err),
			zap.String("route_table_id", routeTable),
		)
		return "", fmt.Errorf("%w: %w",
			errFailedToDeriveVpcIdFromRouteTable, err,
		)
	}

	if len(out.RouteTables) == 0 {
		return "", fmt.Errorf("%w: route-table not found: %s",
			errFailedToDeriveVpcIdFromRouteTable, routeTable,
		)
	}

	rt := out.RouteTables[0]
	if rt.VpcId == nil {
		return "", fmt.Errorf("%w: route-table has not vpc id: %s",
			errFailedToDeriveVpcIdFromRouteTable, routeTable,
		)
	}

	return *rt.VpcId, nil
}

func (cli *Client) FindRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
) ([]*awstypes.Route, error) {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Describing AWS route-table...",
		zap.String("route_table_id", routeTable),
	)

	out, err := cli.ec2.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTable},
	})
	if err != nil {
		l.Error("Failed to describe AWS route-table",
			zap.Error(err),
			zap.String("route_table_id", routeTable),
		)
		return nil, err
	}

	if len(out.RouteTables) == 0 {
		return nil, fmt.Errorf("%w: %s",
			errRouteTableDoesNotExist, routeTable,
		)
	}

	routes := make([]*awstypes.Route, 0, 1)
	for _, rts := range out.RouteTables {
		for _, route := range rts.Routes {
			if aws.ToString(route.DestinationCidrBlock) == cidr {
				routes = append(routes, &route)
			}
		}
	}

	return routes, nil
}

func (cli *Client) UpdateRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Replacing route in AWS route-table...",
		zap.String("destination_cidr_block", cidr),
		zap.String("network_interface_id", networkInterfaceID),
		zap.String("route_table_id", routeTable),
	)

	_, err := cli.ec2.ReplaceRoute(ctx, &ec2.ReplaceRouteInput{
		RouteTableId:         aws.String(routeTable),
		DestinationCidrBlock: aws.String(cidr),
		NetworkInterfaceId:   aws.String(networkInterfaceID),
	})
	if err != nil {
		l.Error("Failed to replace route in AWS route-table",
			zap.Error(err),
			zap.String("cidr", cidr),
			zap.String("network_interface_id", networkInterfaceID),
			zap.String("route_table_id", routeTable),
		)
	}
	return err
}

func (cli *Client) CreateRoute(
	ctx context.Context,
	routeTable string,
	cidr string,
	networkInterfaceID string,
) error {
	l := logutils.LoggerFromContext(ctx)

	l.Debug("Creating route in AWS route-table...",
		zap.String("destination_cidr_block", cidr),
		zap.String("network_interface_id", networkInterfaceID),
		zap.String("route_table_id", routeTable),
	)

	_, err := cli.ec2.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         aws.String(routeTable),
		DestinationCidrBlock: aws.String(cidr),
		NetworkInterfaceId:   aws.String(networkInterfaceID),
	})
	if err != nil {
		l.Error("Failed to create route in AWS route-table",
			zap.Error(err),
			zap.String("destination_cidr_block", cidr),
			zap.String("network_interface_id", networkInterfaceID),
			zap.String("route_table_id", routeTable),
		)
	}
	return err
}

func (cli *Client) DeleteRoute(
	ctx context.Context,
	routeTable string,
	route *awstypes.Route,
) error {
	if route == nil {
		return nil
	}

	l := logutils.LoggerFromContext(ctx)

	l.Debug("Deleting route in AWS route-table...",
		zap.String("destination_cidr_block", aws.ToString(route.DestinationCidrBlock)),
		zap.String("destination_ipv6_cidr_block", aws.ToString(route.DestinationIpv6CidrBlock)),
		zap.String("destination_prefix_list_id", aws.ToString(route.DestinationPrefixListId)),
		zap.String("route_table", routeTable),
	)

	_, err := cli.ec2.DeleteRoute(ctx, &ec2.DeleteRouteInput{
		RouteTableId:             aws.String(routeTable),
		DestinationCidrBlock:     route.DestinationCidrBlock,
		DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
		DestinationPrefixListId:  route.DestinationPrefixListId,
	})
	if err != nil {
		l.Error("Failed to delete route in AWS route-table",
			zap.String("destination_cidr_block", aws.ToString(route.DestinationCidrBlock)),
			zap.String("destination_ipv6_cidr_block", aws.ToString(route.DestinationIpv6CidrBlock)),
			zap.String("destination_prefix_list_id", aws.ToString(route.DestinationPrefixListId)),
			zap.String("route_table", routeTable),
		)
	}
	return err
}
