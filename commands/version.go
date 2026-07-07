package commands

import (
	"context"
	"fmt"

	"github.com/doron-cohen/klee/version"
	"github.com/urfave/cli/v3"
)

// VersionCommand returns the built-in version subcommand.
func VersionCommand(ver string) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "print version information",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "short",
				Usage: "print version number only",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("short") {
				fmt.Println(version.Version)
			} else {
				fmt.Println(ver)
			}
			return nil
		},
	}
}
