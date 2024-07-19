package config

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

type ReconcileInterfaceActivate struct {
	Reapply *ReconcileReapply `yaml:"reapply"`

	Script types.Script `yaml:"script"`
}

func (r *ReconcileInterfaceActivate) PostLoad(ctx context.Context) error {
	if r.Reapply == nil {
		r.Reapply = &ReconcileReapply{}
	}

	if err := r.Reapply.PostLoad(ctx); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileInterfaceActivate) Validate(ctx context.Context) error {
	if err := r.Reapply.Validate(ctx); err != nil {
		return err
	}

	return nil
}
