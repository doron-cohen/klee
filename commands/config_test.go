package commands_test

import (
	"context"
	"testing"

	"github.com/doron-cohen/klee/commands"
	"github.com/doron-cohen/klee/kleetest"
	"github.com/urfave/cli/v3"
)

type demoConfig struct {
	Host string `yaml:"host"`
}

// cmdRunner adapts a cli.Command to kleetest.Runner.
type cmdRunner struct{ cmd *cli.Command }

func (r cmdRunner) Run(ctx context.Context, args []string) int {
	if err := r.cmd.Run(ctx, args); err != nil {
		return 1
	}
	return 0
}

// An App whose LoadConfig never ran has no sources, and config print must
// still work rather than emit an empty heading.
func TestConfigPrintWithoutSources(t *testing.T) {
	cfg := commands.ConfigCommand(func(context.Context) any {
		return &demoConfig{Host: "example"}
	}, nil)

	result := kleetest.Run(t, cmdRunner{&cli.Command{Commands: []*cli.Command{cfg}}}, "config", "print")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Equals(t, "host: example\n")
	result.Stderr.Empty(t)
}
