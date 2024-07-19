package config

import (
	"time"

	"github.com/flashbots/vpnham/types"
)

type Scripts struct {
	Timeout time.Duration `yaml:"timeout"`

	BridgeActivate      types.Script `yaml:"bridge_activate"`
	InterfaceActivate   types.Script `yaml:"interface_activate"`
	InterfaceDeactivate types.Script `yaml:"interface_deactivate"`
}
