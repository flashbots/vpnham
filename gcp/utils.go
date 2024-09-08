package gcp

import "strings"

func (cli *Client) NormaliseNetworkID(networkID string) string {
	prefix := "projects/" + cli.projectNumber + "/networks/"
	networkID = strings.TrimPrefix(networkID, prefix)
	networkID = "https://www.googleapis.com/compute/v1/projects/" + cli.projectID + "/global/networks/" + networkID
	return networkID
}

func (cli *Client) NormaliseInstanceName(instanceName string) string {
	prefix := "projects/" + cli.projectID + "/zones/" + cli.zone + "/instances/"
	instanceName = strings.TrimPrefix(instanceName, prefix)
	instanceName = "https://www.googleapis.com/compute/v1/" + prefix + instanceName
	return instanceName
}
