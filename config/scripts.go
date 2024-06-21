package config

import (
	"github.com/flashbots/vpnham/types"
)

type Scripts struct {
	BridgeActivate      types.Script `yaml:"bridge_activate"`
	InterfaceActivate   types.Script `yaml:"interface_activate"`
	InterfaceDeactivate types.Script `yaml:"interface_deactivate"`
}
