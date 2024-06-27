package executor

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

type job struct {
	name   string
	script *types.Script
}

func (ex *Executor) loop(ctx context.Context) {
	exhaust := make(chan job, 1)

	for {
		select {
		case job := <-exhaust:
			ex.execute(ctx, job)
		case job := <-ex.next:
			ex.execute(ctx, job)
		case <-ex.stop:
			return
		}

		ex.mxQueue.Lock()
		if len(ex.queue) > 0 {
			exhaust <- ex.queue[0]
			ex.queue = ex.queue[1:]
		}
		ex.mxQueue.Unlock()
	}
}

func (ex *Executor) schedule(job job) {
	ex.mxQueue.Lock()
	defer ex.mxQueue.Unlock()

	select {
	case ex.next <- job:
		break
	default:
		ex.queue = append(ex.queue, job)
	}
}
