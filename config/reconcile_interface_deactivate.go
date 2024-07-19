package config

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

type ReconcileInterfaceDeactivate struct {
	Script types.Script `yaml:"script"`
}

func (r *ReconcileInterfaceDeactivate) PostLoad(ctx context.Context) error {
	return nil
}

func (r *ReconcileInterfaceDeactivate) Validate(ctx context.Context) error {
	return nil
}
