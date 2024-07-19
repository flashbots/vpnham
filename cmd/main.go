package main

import (
	"fmt"
	"os"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/logutils"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	version = "development"
)

const (
	envPrefix = "VPNHAM_"
)

func main() {
	cfg := &config.Config{
		Log: &config.Log{},
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Destination: &cfg.Log.Level,
			EnvVars:     []string{envPrefix + "LOG_LEVEL"},
			Name:        "log-level",
			Usage:       "logging level",
			Value:       config.DefaultLogLevel,
		},

		&cli.StringFlag{
			Destination: &cfg.Log.Mode,
			EnvVars:     []string{envPrefix + "LOG_MODE"},
			Name:        "log-mode",
			Usage:       "logging mode",
			Value:       config.DefaultLogMode,
		},
	}

	commands := []*cli.Command{
		CommandServe(cfg),
		CommandHelp(cfg),
	}

	app := &cli.App{
		Name:    "vpnham",
		Usage:   "Monitor highly-available VPN tunnels",
		Version: version,

		Flags:          flags,
		Commands:       commands,
		DefaultCommand: commands[0].Name,

		Before: func(_ *cli.Context) error {
			// setup logger
			l, err := logutils.NewLogger(cfg.Log.Mode, cfg.Log.Level)
			if err != nil {
				return err
			}
			zap.ReplaceGlobals(l)

			return nil
		},

		Action: func(clictx *cli.Context) error {
			return cli.ShowAppHelp(clictx)
		},
	}

	defer func() {
		zap.L().Sync() //nolint:errcheck
	}()
	if err := app.Run(os.Args); err != nil {
		if err.Error() == "flag: help requested" {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "\nFailed with error:\n\n%s\n\n", err.Error())
		os.Exit(1)
	}
}
