package config

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

type ReconcileBridgeActivate struct {
	BridgeName          string   `yaml:"-"`
	BridgeInterface     string   `yaml:"-"`
	SecondaryInterfaces []string `yaml:"-"`

	Reapply *ReconcileReapply `yaml:"reapply"`

	AWS *ReconcileBridgeActivateAWS `yaml:"aws"`
	GCP *ReconcileBridgeActivateGCP `yaml:"gcp"`

	Script types.Script `yaml:"script"`
}

func (r *ReconcileBridgeActivate) PostLoad(ctx context.Context) error {
	if r.Reapply == nil {
		r.Reapply = &ReconcileReapply{}
	}

	if err := r.Reapply.PostLoad(ctx); err != nil {
		return err
	}

	if r.AWS != nil {
		r.AWS.BridgeName = r.BridgeName
		r.AWS.BridgeInterface = r.BridgeInterface
		r.AWS.SecondaryInterfaces = r.SecondaryInterfaces

		if err := r.AWS.PostLoad(ctx); err != nil {
			return err
		}
	}

	if r.GCP != nil {
		r.GCP.BridgeName = r.BridgeName
		r.GCP.BridgeInterface = r.BridgeInterface
		r.GCP.SecondaryInterfaces = r.SecondaryInterfaces

		if err := r.GCP.PostLoad(ctx); err != nil {
			return err
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
