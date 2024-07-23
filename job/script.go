package job

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/metrics"
	"github.com/flashbots/vpnham/types"
	"github.com/flashbots/vpnham/utils"
	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type runScript struct {
	name    string
	timeout time.Duration

	script types.Script
}

func RunScript(
	name string,
	timeout time.Duration,
	script types.Script,
) Job {
	return &runScript{
		name:    name,
		timeout: timeout,
		script:  script,
	}
}

func (j *runScript) Name() string {
	return j.name
}

func (j *runScript) Execute(ctx context.Context) error {
	l := logutils.LoggerFromContext(ctx)

	errs := []error{}
	for step, _cmd := range j.script {
		if len(_cmd) == 0 {
			continue
		}

		strCmd := strings.Join(_cmd, " ")

		l.Debug("Executing command",
			zap.String("command", strCmd),
		)

		ctx, cancel := context.WithTimeout(ctx, j.timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, _cmd[0], _cmd[1:]...)

		stdout := &strings.Builder{}
		cmd.Stdout = stdout

		stderr := &strings.Builder{}
		cmd.Stderr = stderr

		cmd.Env = os.Environ()

		start := time.Now()
		err := utils.WithTimeout(ctx, j.timeout, func(ctx context.Context) error {
			return cmd.Run()
		})
		duration := time.Since(start)

		if err != nil {
			metrics.Errors.Add(ctx, 1, otelapi.WithAttributes(
				attribute.String(metrics.LabelErrorScope, "job_"+j.name),
			))
			errs = append(errs, err)
		}

		l.Info("Executed command",
			zap.String("script", j.name),
			zap.Int("step", step),
			zap.String("command", strCmd),
			zap.Int64("duration_us", duration.Microseconds()),

			zap.String("stderr", strings.TrimSpace(stderr.String())),
			zap.String("stdout", strings.TrimSpace(stdout.String())),

			zap.Error(err),
		)
	}

	switch len(errs) {
	default:
		return errors.Join(errs...)
	case 1:
		return errs[0]
	case 0:
		return nil
	}
}
