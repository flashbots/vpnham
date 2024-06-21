package executor

import (
	"context"

	"github.com/flashbots/vpnham/types"
)

func (ex *Executor) loop(ctx context.Context) {
	exhaust := make(chan *types.Script, 1)

	for {
		select {
		case script := <-exhaust:
			ex.execute(ctx, script)
		case script := <-ex.next:
			ex.execute(ctx, script)
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

func (ex *Executor) schedule(script *types.Script) {
	ex.mxQueue.Lock()
	defer ex.mxQueue.Unlock()

	select {
	case ex.next <- script:
		break
	default:
		ex.queue = append(ex.queue, script)
	}
}
