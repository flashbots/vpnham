package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/flashbots/vpnham/monitor"
	"github.com/flashbots/vpnham/types"
)

type TunnelInterface struct {
	Name string

	Role      types.Role    `yaml:"role"`
	Addr      types.Address `yaml:"addr"`
	ProbeAddr types.Address `yaml:"probe_addr"`

	ThresholdDown int `yaml:"threshold_down"`
	ThresholdUp   int `yaml:"threshold_up"`
}

var (
	errTunnelInterfaceAddrIsInvalid              = errors.New("tunnel interface addr is invalid")
	errTunnelInterfaceProbeAddrIsInvalid         = errors.New("tunnel interface probe addr is invalid")
	errTunnelInterfaceRoleIsInvalid              = errors.New("tunnel interface role is invalid")
	errTunnelInterfaceStatusThresholdsAreInvalid = errors.New("tunnel interface status thresholds are invalid")
)

func (ifs *TunnelInterface) PostLoad(ctx context.Context) error {
	if ifs.ThresholdDown == 0 {
		ifs.ThresholdDown = DefaultThresholdDown
	}

	if ifs.ThresholdUp == 0 {
		ifs.ThresholdUp = DefaultThresholdUp
	}

	return nil
}

func (ifs *TunnelInterface) Validate(ctx context.Context) error {
	if err := ifs.Role.Validate(); err != nil {
		return fmt.Errorf("%s: %w: %w",
			ifs.Name, errTunnelInterfaceRoleIsInvalid, err,
		)
	}

	if err := ifs.Addr.Validate(); err != nil {
		return fmt.Errorf("%s: %w: %w",
			ifs.Name, errTunnelInterfaceAddrIsInvalid, err,
		)
	}

	if err := ifs.ProbeAddr.Validate(); err != nil {
		return fmt.Errorf("%s: %w: %w",
			ifs.Name, errTunnelInterfaceProbeAddrIsInvalid, err,
		)
	}

	if _, err := monitor.New(ifs.ThresholdDown, ifs.ThresholdUp); err != nil {
		return fmt.Errorf("%s: %w: %w",
			ifs.Name, errTunnelInterfaceStatusThresholdsAreInvalid, err,
		)
	}

	return nil
}
