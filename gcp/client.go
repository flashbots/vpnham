package gcp

import (
	"context"
	"errors"

	gce "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
)

type Client struct {
	instanceName  string
	projectID     string
	projectNumber string
	zone          string

	routes *gce.RoutesClient
}

var (
	errGCPNotOnGCE = errors.New("didn't detect google compute engine")
)

func NewClient(ctx context.Context) (*Client, error) {
	if !metadata.OnGCE() {
		return nil, errGCPNotOnGCE
	}

	projectID, err := metadata.ProjectIDWithContext(ctx)
	if err != nil {
		return nil, err
	}

	projectNumber, err := metadata.NumericProjectIDWithContext(ctx)
	if err != nil {
		return nil, err
	}

	zone, err := metadata.ZoneWithContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceName, err := metadata.InstanceNameWithContext(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := gce.NewRoutesRESTClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		instanceName:  instanceName,
		projectID:     projectID,
		projectNumber: projectNumber,
		zone:          zone,
		routes:        routes,
	}, nil
}
