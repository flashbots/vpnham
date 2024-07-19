package reconciler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/types"
	"go.uber.org/zap"
)

type job struct {
	name   string
	script *types.Script
}

func (r *Reconciler) runLoop(
	ctx context.Context,
) {
	exhaust := make(chan job, 1)

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
	job job,
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
	job job,
) {
	l := logutils.LoggerFromContext(ctx)

	for step, _cmd := range *job.script {
		if len(_cmd) == 0 {
			continue
		}

		strCmd := strings.Join(_cmd, " ")

		l.Debug("Executing command",
			zap.String("command", strCmd),
		)

		ctx, cancel := context.WithTimeout(ctx, r.cfg.ScriptsTimeout)
		defer cancel()

		start := time.Now()
		cmd := exec.CommandContext(ctx, _cmd[0], _cmd[1:]...)
		duration := time.Since(start)

		stdout := &strings.Builder{}
		stderr := &strings.Builder{}

		cmd.Env = os.Environ()
		cmd.Stderr = stderr
		cmd.Stdout = stdout

		err := cmd.Run()
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("timed out after %v: %w", time.Since(start), err)
		}

		l.Info("Executed command",
			zap.String("script", job.name),
			zap.Int("step", step),
			zap.String("command", strCmd),
			zap.Int64("duration_us", duration.Microseconds()),

			zap.String("stderr", strings.TrimSpace(stderr.String())),
			zap.String("stdout", strings.TrimSpace(stdout.String())),

			zap.Error(err),
		)
	}
}
