package gcp

import (
	"context"
)

func (cli *Client) ProjectID(ctx context.Context) (string, error) {
	return cli.projectID, nil
}

func (cli *Client) InstanceName(ctx context.Context) (string, error) {
	return cli.instanceName, nil
}
