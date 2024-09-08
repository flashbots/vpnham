package aws

import (
	"context"
)

func (cli *Client) Region(ctx context.Context) (string, error) {
	return cli.region, nil
}
