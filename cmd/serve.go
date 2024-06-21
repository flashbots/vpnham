package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/flashbots/vpnham/config"
	"github.com/flashbots/vpnham/server"
	"github.com/urfave/cli/v2"
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

		Before: func(ctx *cli.Context) error {
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
			cfgServer.PostLoad()
			if err := cfgServer.Validate(); err != nil {
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
