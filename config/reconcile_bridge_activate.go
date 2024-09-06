package config

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

type ReconcileBridgeActivate struct {
	BridgeInterface     string   `yaml:"-"`
	SecondaryInterfaces []string `yaml:"-"`

	Reapply *ReconcileReapply `yaml:"reapply"`

	AWS    *ReconcileBridgeActivateAWS `yaml:"aws"`
	Script types.Script                `yaml:"script"`
}

func (r *ReconcileBridgeActivate) PostLoad(ctx context.Context) error {
	if r.Reapply == nil {
		r.Reapply = &ReconcileReapply{}
	}

	if err := r.Reapply.PostLoad(ctx); err != nil {
		return err
	}

	if r.AWS != nil {
		r.AWS.BridgeInterface = r.BridgeInterface
		r.AWS.SecondaryInterfaces = r.SecondaryInterfaces

		if err := r.AWS.PostLoad(ctx); err != nil {
			return err
		}
		if r.AWS.Timeout == 0 {
			r.AWS.Timeout = DefaultAWSTimeout
		}
	}

	return nil
}

func (r *ReconcileBridgeActivate) Validate(ctx context.Context) error {
	if err := r.Reapply.Validate(ctx); err != nil {
		return err
	}

	return nil
}
