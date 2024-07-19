package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/logutils"
	"github.com/flashbots/vpnham/server"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	errConfigFailedToRead = errors.New("failed to read config")
	errConfigIsInvalid    = errors.New("invalid config")
)

func CommandServe(cfg *config.Config) *cli.Command {
	configFile := ""

	serverFlags := []cli.Flag{
		&cli.StringFlag{
			Destination: &configFile,
			EnvVars:     []string{envPrefix + "_CONFIG"},
			Name:        "config",
			Usage:       "a `path` to the configuration file",
			Value:       config.DefaultConfigFile,
		},
	}

	flags := slices.Concat(
		serverFlags,
	)

	return &cli.Command{
		Name:  "serve",
		Usage: "run vpnham server",
		Flags: flags,

		Before: func(_ *cli.Context) error {
			l := zap.L()
			ctx := logutils.ContextWithLogger(context.Background(), l)

			if _, err := os.Stat(configFile); err != nil {
				return fmt.Errorf("%w: %w", errConfigFailedToRead, err)
			}
			b, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("%w: %w", errConfigFailedToRead, err)
			}
			cfgServer := &config.Server{}
			if err := yaml.UnmarshalStrict(b, cfgServer); err != nil {
				return fmt.Errorf("%w: %w", errConfigFailedToRead, err)
			}
			if err := cfgServer.PostLoad(ctx); err != nil {
				return err
			}
			if err := cfgServer.Validate(ctx); err != nil {
				return fmt.Errorf("%w: %w", errConfigIsInvalid, err)
			}
			cfg.Server = cfgServer
			return nil
		},

		Action: func(_ *cli.Context) error {
			s, err := server.New(cfg)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}
}
