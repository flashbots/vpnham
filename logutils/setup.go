package logutils

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	errLoggerFailedToBuild = errors.New("failed to build the logger")
	errLoggerInvalidLevel  = errors.New("invalid log-level")
	errLoggerInvalidMode   = errors.New("invalid log-mode")
)

func NewLogger(mode, level string) (
	*zap.Logger, error,
) {
	var config zap.Config
	switch strings.ToLower(mode) {
	case "dev":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeCaller = nil
	case "prod":
		config = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("%w: %s",
			errLoggerInvalidMode, mode,
		)
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w",
			errLoggerInvalidLevel, level, err,
		)
	}
	config.Level = logLevel

	l, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errLoggerFailedToBuild, err,
		)
	}

	return l, nil
}
