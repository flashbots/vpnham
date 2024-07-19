package config

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

type ReconcileReapply struct {
	maxIterations int

	// -

	InitialDelay time.Duration `yaml:"initial_delay"`
	MaximumDelay time.Duration `yaml:"maximum_delay"`

	Factor float64 `yaml:"factor"`
}

var (
	errReconcileReapplyFactorIsInvalid       = errors.New("invalid reapply factor")
	errReconcileReapplyInitialDelayIsInvalid = errors.New("invalid initial reapply delay")
	errReconcileReapplyMaximumDelayIsInvalid = errors.New("invalid maximum reapply delay")
)

func (rr *ReconcileReapply) PostLoad(ctx context.Context) error {
	if rr.Factor == 0.0 {
		rr.Factor = DefaultReapplyFactor
	}

	if rr.MaximumDelay == 0.0 && rr.InitialDelay != 0.0 {
		rr.MaximumDelay = rr.InitialDelay
	}

	if rr.InitialDelay == 0.0 && rr.MaximumDelay != 0.0 {
		rr.InitialDelay = rr.MaximumDelay
	}

	return nil
}

func (rr *ReconcileReapply) Validate(ctx context.Context) error {
	if rr.InitialDelay != 0 && rr.InitialDelay < time.Second {
		return fmt.Errorf("%w: expected >= 1s, got %s",
			errReconcileReapplyInitialDelayIsInvalid, rr.InitialDelay,
		)
	}

	if rr.MaximumDelay < rr.InitialDelay {
		return fmt.Errorf("%w: expected >= %s, got %s",
			errReconcileReapplyMaximumDelayIsInvalid, rr.InitialDelay, rr.MaximumDelay,
		)
	}

	if rr.Factor < 1.0 {
		return fmt.Errorf("%w: expected >= 1.0, got %f",
			errReconcileReapplyFactorIsInvalid, rr.Factor,
		)
	}

	if rr.Factor == 1.00 && rr.InitialDelay != rr.MaximumDelay {
		return fmt.Errorf("%w: expected > 1.0, got %f",
			errReconcileReapplyFactorIsInvalid, rr.Factor,
		)
	}

	return nil
}

func (rr *ReconcileReapply) Enabled() bool {
	if rr == nil {
		return false
	}

	return rr.InitialDelay > 0
}

func (rr *ReconcileReapply) DelayOnIteration(iteration int) time.Duration {
	if rr.maxIterations == 0 {
		if rr.Factor == 1.0 {
			rr.maxIterations = 1
		} else {
			ratio := float64(rr.MaximumDelay) / float64(rr.InitialDelay)
			maxIterations := int(math.Ceil(math.Log(ratio) / math.Log(rr.Factor)))
			if maxIterations == 0 {
				maxIterations = 1
			}
			rr.maxIterations = maxIterations
		}
	}

	if iteration >= rr.maxIterations {
		return rr.MaximumDelay
	}

	delay := time.Second * time.Duration(math.Pow(rr.Factor, float64(iteration))*rr.InitialDelay.Seconds())
	if delay > rr.MaximumDelay {
		delay = rr.MaximumDelay
	}
	return delay
}
