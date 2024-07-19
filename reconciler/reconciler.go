package reconciler

import (
	"context"
	"sync"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/logutils"
)

type Reconciler struct {
	name string

	cfg *config.Reconcile

	queue   []job
	mxQueue sync.Mutex

	next chan job
	stop chan struct{}
}

func New(name string, cfg *config.Reconcile) (*Reconciler, error) {
	r := &Reconciler{
		name: name,

		cfg: cfg,

		queue: make([]job, 0, 1),

		next: make(chan job),
		stop: make(chan struct{}, 1),
	}

	return r, nil
}

func (r *Reconciler) Run(ctx context.Context, failureSink chan<- error) {
	l := logutils.LoggerFromContext(ctx)

	l.Info("VPN HA-monitor reconciler is going up...")

	go func() {
		r.runLoop(ctx)

		l.Info("VPN HA-monitor reconciler is down")
	}()
}

func (r *Reconciler) Stop(ctx context.Context) {
	r.stop <- struct{}{}
}
