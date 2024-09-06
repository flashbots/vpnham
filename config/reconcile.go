package config

import (
	"context"
	"time"
)

type Reconcile struct {
	BridgeInterface     string   `yaml:"-"`
	SecondaryInterfaces []string `yaml:"-"`

	ScriptsTimeout time.Duration `yaml:"scripts_timeout"`

	BridgeActivate      *ReconcileBridgeActivate      `yaml:"bridge_activate"`
	InterfaceActivate   *ReconcileInterfaceActivate   `yaml:"interface_activate"`
	InterfaceDeactivate *ReconcileInterfaceDeactivate `yaml:"interface_deactivate"`
}

func (r *Reconcile) PostLoad(ctx context.Context) error {
	if r.ScriptsTimeout == 0 {
		r.ScriptsTimeout = DefaultScriptsTimeout
	}

	{ // bridge_activate
		if r.BridgeActivate == nil {
			r.BridgeActivate = &ReconcileBridgeActivate{}
		}
		r.BridgeActivate.BridgeInterface = r.BridgeInterface
		r.BridgeActivate.SecondaryInterfaces = r.SecondaryInterfaces

		if err := r.BridgeActivate.PostLoad(ctx); err != nil {
			return err
		}
	}

	{ // interface_activate
		if r.InterfaceActivate == nil {
			r.InterfaceActivate = &ReconcileInterfaceActivate{}
		}

		if err := r.InterfaceActivate.PostLoad(ctx); err != nil {
			return err
		}
	}

	{ // interface_deactivate
		if r.InterfaceDeactivate == nil {
			r.InterfaceDeactivate = &ReconcileInterfaceDeactivate{}
		}

		if err := r.InterfaceDeactivate.PostLoad(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconcile) Validate(ctx context.Context) error {
	if err := r.BridgeActivate.Validate(ctx); err != nil {
		return err
	}

	if err := r.InterfaceActivate.Validate(ctx); err != nil {
		return err
	}

	if err := r.InterfaceDeactivate.Validate(ctx); err != nil {
		return err
	}

	return nil
}
