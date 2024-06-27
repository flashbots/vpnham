package executor

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

func (ex *Executor) render(source *types.Script, params map[string]string) *types.Script {
	resScript := make(types.Script, 0, len(*source))
	for _, cmd := range *source {
		resCmd := make(types.Command, 0, len(cmd))
		for _, elem := range cmd {
			resElem := elem
			for placeholder, value := range params {
				resElem = strings.ReplaceAll(resElem, "${"+placeholder+"}", value)
			}
			resCmd = append(resCmd, resElem)
		}
		resScript = append(resScript, resCmd)
	}

	return &resScript
}

func (ex *Executor) execute(ctx context.Context, job job) {
	l := logutils.LoggerFromContext(ctx)

	for step, _cmd := range *job.script {
		if len(_cmd) == 0 {
			continue
		}

		strCmd := strings.Join(_cmd, " ")

		l.Debug("Executing command",
			zap.String("command", strCmd),
		)

		ctx, cancel := context.WithTimeout(ctx, ex.timeout)
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
			err = fmt.Errorf("timed out after %v: %w", ex.timeout, err)
		}

		l.Info("Executed command",
			zap.String("script", job.name),
			zap.Int("step", step),
			zap.String("command", strCmd),
			zap.Duration("duration", duration),

			zap.String("stderr", strings.TrimSpace(stderr.String())),
			zap.String("stdout", strings.TrimSpace(stdout.String())),

			zap.Error(err),
		)
	}
}
