package config

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/flashbots/vpnham/monitor"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"github.com/google/uuid"
)

type Bridge struct {
	Name string    `yaml:"-"`
	UUID uuid.UUID `yaml:"-"`

	Role types.Role `yaml:"role"`

	BridgeInterface string     `yaml:"bridge_interface"`
	PeerCIDR        types.CIDR `yaml:"peer_cidr"`

	StatusAddr                 types.Address `yaml:"status_addr"`
	PartnerStatusURL           string        `yaml:"partner_status_url"`
	PartnerStatusTimeout       time.Duration `yaml:"partner_status_timeout"`
	PartnerStatusThresholdDown int           `yaml:"partner_status_threshold_down"`
	PartnerStatusThresholdUp   int           `yaml:"partner_status_threshold_up"`

	ProbeInterval time.Duration  `yaml:"probe_interval"`
	ProbeLocation types.Location `yaml:"probe_location"`

	TunnelInterfaces map[string]*TunnelInterface `yaml:"tunnel_interfaces"`

	ScriptsTimeout time.Duration `yaml:"scripts_timeout"`
	Scripts        *Scripts      `yaml:"scripts"`
}

var (
	errBridgeActiveTunnelInterfacesCountIsInvalid = errors.New("bridge has invalid count of active interfaces configured (must be only 1)")
	errBridgeInterfaceIsInvalid                   = errors.New("bridge interface is invalid")
	errBridgePartnerStatusThresholdsAreInvalid    = errors.New("bridge partner status thresholds are invalid")
	errBridgePartnerStatusURLIsInvalid            = errors.New("bridge partner status url is invalid")
	errBridgePeerCIDRIsInvalid                    = errors.New("bridge peer cidr is invalid")
	errBridgeRoleIsInvalid                        = errors.New("bridge role is invalid")
	errBridgeStatusAddrIsInvalid                  = errors.New("bridge status addr is invalid")
	errBridgeTunnelInterfaceIsInvalid             = errors.New("bridge tunnel interface is invalid")
)

func (b *Bridge) Validate() error {
	{ // role
		if err := b.Role.Validate(); err != nil {
			return fmt.Errorf("%w: %q",
				errBridgeRoleIsInvalid, err,
			)
		}
	}

	{ // bridge_interface
		if _, _, err := utils.GetInterfaceIPs(b.BridgeInterface); err != nil {
			return fmt.Errorf("%w: %w",
				errBridgeInterfaceIsInvalid, err,
			)
		}
	}

	{ // peer_cidr
		if err := b.PeerCIDR.Validate(); err != nil {
			return fmt.Errorf("%w: %w",
				errBridgePeerCIDRIsInvalid, err,
			)
		}
	}

	{ // status_addr
		if err := b.StatusAddr.Validate(); err != nil {
			return fmt.Errorf("%w: %w",
				errBridgeStatusAddrIsInvalid, err,
			)
		}
	}

	{ // partner_status_url
		if _, err := url.Parse(b.PartnerStatusURL); err != nil {
			return fmt.Errorf("%w: %w",
				errBridgePartnerStatusURLIsInvalid, err,
			)
		}
	}

	{ // partner_status_threshold_down, partner_status_threshold_up
		if _, err := monitor.New(b.PartnerStatusThresholdDown, b.PartnerStatusThresholdUp); err != nil {
			return fmt.Errorf("%w: %w",
				errBridgePartnerStatusThresholdsAreInvalid, err,
			)
		}
	}

	// probe_interval is validated at un-marshalling

	// probe_location is validated at un-marshalling

	{ // tunnel_interfaces
		activeInterfacesCount := 0
		for ifsName, ifs := range b.TunnelInterfaces {
			if _, _, err := utils.GetInterfaceIPs(ifsName); err != nil {
				return fmt.Errorf("%w: %w",
					errBridgeTunnelInterfaceIsInvalid, err,
				)
			}

			if err := ifs.Validate(); err != nil {
				return fmt.Errorf("%w: %w",
					errBridgeTunnelInterfaceIsInvalid, err,
				)
			}

			if ifs.Role == types.RoleActive {
				activeInterfacesCount++
			}
		}
		if activeInterfacesCount != 1 {
			return fmt.Errorf("%w: %d",
				errBridgeActiveTunnelInterfacesCountIsInvalid, activeInterfacesCount,
			)
		}
	}

	return nil
}

func (b *Bridge) TunnelInterfacesCount() int {
	return len(b.TunnelInterfaces)
}
