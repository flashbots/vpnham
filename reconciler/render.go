package reconciler

import (
	"strings"

	"github.com/flashbots/vpnham/event"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
)

const (
	placeholderProto = "proto"

	placeholderBridgeInterface      = "bridge_interface"
	placeholderBridgeInterfaceIP    = "bridge_interface_ip"
	placeholderBridgePeerCIDR       = "bridge_peer_cidr"
	placeholderBridgeExtraPeerCIDRs = "bridge_extra_peer_cidrs"

	placeholderTunnelInterface      = "tunnel_interface"
	placeholderTunnelInterfaceIP    = "tunnel_interface_ip"
	placeholderTunnelInterfaceProto = "tunnel_interface_proto"
)

func (r *Reconciler) renderPlaceholders(e event.Event) (map[string]string, error) {
	placeholders := map[string]string{}
	var err error

	if e, ok := e.(event.BridgeEvent); ok {
		placeholders[placeholderBridgeInterface] = e.EvtBridgeInterface()

		cidrs := e.EvtBridgePeerCIDRs()
		if len(cidrs) > 0 {
			bridgePeerCIDR := cidrs[0]
			placeholders[placeholderBridgePeerCIDR] = bridgePeerCIDR.String()

			ipv4 := bridgePeerCIDR.IsIPv4()
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

		if len(cidrs) > 1 {
			extraPeerCIDRs := make([]string, 0, len(cidrs)-1)
			for idx := 1; idx < len(cidrs); idx++ {
				extraPeerCIDRs = append(extraPeerCIDRs, cidrs[idx].String())
			}

			placeholders[placeholderBridgeExtraPeerCIDRs] = strings.Join(extraPeerCIDRs, ",")
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
