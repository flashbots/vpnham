package main

import (
	"github.com/flashbots/vpnham/config"
	"github.com/urfave/cli/v2"
)

func CommandHelp(_ *config.Config) *cli.Command {
	return &cli.Command{
		Usage: "show a list of commands or help for one command",
		Name:  "help",

		Action: func(clictx *cli.Context) error {
			return cli.ShowAppHelp(clictx)
		},
	}
}
