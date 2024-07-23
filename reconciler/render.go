package reconciler

import (
	"strings"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
)

const (
	placeholderProto = "proto"

	placeholderBridgeInterface   = "bridge_interface"
	placeholderBridgeInterfaceIP = "bridge_interface_ip"
	placeholderBridgePeerCIDR    = "bridge_peer_cidr"

	placeholderTunnelInterface      = "tunnel_interface"
	placeholderTunnelInterfaceIP    = "tunnel_interface_ip"
	placeholderTunnelInterfaceProto = "tunnel_interface_proto"
)

func (r *Reconciler) renderPlaceholders(e event.Event) (map[string]string, error) {
	placeholders := map[string]string{}
	var err error

	if e, ok := e.(event.BridgeEvent); ok {
		placeholders[placeholderBridgeInterface] = e.EvtBridgeInterface()
		placeholders[placeholderBridgePeerCIDR] = e.EvtBridgePeerCIDR().String()

		ipv4 := e.EvtBridgePeerCIDR().IsIPv4()
		if ipv4 {
			placeholders[placeholderProto] = "4"
		} else {
			placeholders[placeholderProto] = "6"
		}

		placeholders[placeholderBridgeInterfaceIP], err = utils.GetInterfaceIP(e.EvtBridgeInterface(), ipv4)
		if err != nil {
			return nil, err
		}

		if e, ok := e.(event.TunnelInterfaceEvent); ok {
			placeholders[placeholderTunnelInterface] = e.EvtTunnelInterface()

			placeholders[placeholderTunnelInterfaceIP], err = utils.GetInterfaceIP(e.EvtTunnelInterface(), ipv4)
			if err != nil {
				return nil, err
			}
		}
	}

	return placeholders, nil
}

func (r *Reconciler) renderScript(
	source *types.Script,
	params map[string]string,
) types.Script {
	resScript := make(types.Script, 0, len(*source))
	for _, cmd := range *source {
		resCmd := make(types.Command, 0, len(cmd))
		for _, elem := range cmd {
			resElem := elem
			for placeholder, value := range params {
				resElem = strings.ReplaceAll(resElem, "${"+placeholder+"}", value)
			}
			resCmd = append(resCmd, resElem)
		}
		resScript = append(resScript, resCmd)
	}

	return resScript
}
