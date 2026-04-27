package klee

import "github.com/urfave/cli/v3"

var globalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "config",
		Usage: "path to config file",
	},
	&cli.StringFlag{
		Name:  "log-level",
		Usage: "log level (debug, info, warn, error)",
	},
	&cli.StringFlag{
		Name:  "log-format",
		Usage: "log format (pretty, json)",
	},
	&cli.BoolFlag{
		Name:  "no-color",
		Usage: "disable color output",
	},
	&cli.BoolFlag{
		Name:  "json",
		Usage: "output as JSON",
	},
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug output",
	},
	&cli.BoolFlag{
		Name:  "quiet",
		Usage: "suppress non-essential output",
	},
}
