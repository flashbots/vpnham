package reconciler

import (
	"context"
	"time"

	"github.com/flashbots/vpnham/job"
	"github.com/flashbots/vpnham/logutils"
	"go.uber.org/zap"
)

func (r *Reconciler) runLoop(
	ctx context.Context,
) {
	exhaust := make(chan job.Job, 1)

	for {
		select {
		case job := <-exhaust:
			r.executeJob(ctx, job)
		case job := <-r.next:
			r.executeJob(ctx, job)
		case <-r.stop:
			return
		}

		r.mxQueue.Lock()
		if len(r.queue) > 0 {
			exhaust <- r.queue[0]
			r.queue = r.queue[1:]
		}
		r.mxQueue.Unlock()
	}
}

func (r *Reconciler) scheduleJob(
	job job.Job,
) {
	r.mxQueue.Lock()
	defer r.mxQueue.Unlock()

	select {
	case r.next <- job:
		break
	default:
		r.queue = append(r.queue, job)
	}
}

func (r *Reconciler) executeJob(
	ctx context.Context,
	job job.Job,
) {
	l := logutils.LoggerFromContext(ctx)

	start := time.Now()
	err := job.Execute(ctx)
	duration := time.Since(start)

	if err == nil {
		l.Info("Executed job",
			zap.Int64("duration_us", duration.Microseconds()),
			zap.String("job_name", job.GetJobName()),
		)
	} else {
		l.Error("Failed job",
			zap.Error(err),
			zap.Int64("duration_us", duration.Microseconds()),
			zap.String("job_name", job.GetJobName()),
		)
	}
}
