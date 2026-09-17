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

func runConfigCommand(t *testing.T, opts ...commands.ConfigOption) *kleetest.Result {
	t.Helper()
	cfg := commands.ConfigCommand(func(context.Context) any {
		return &demoConfig{Host: "example"}
	}, opts...)
	return kleetest.Run(t, cmdRunner{&cli.Command{Commands: []*cli.Command{cfg}}}, "config", "print")
}

// The v0.2.2 call form takes no options and must still print only the config.
func TestConfigPrintWithoutSources(t *testing.T) {
	result := runConfigCommand(t)
	result.ExitCode.Equals(t, 0)
	result.Stdout.Equals(t, "host: example\n")
	result.Stderr.Empty(t)
}

func TestConfigPrintWithEmptySources(t *testing.T) {
	result := runConfigCommand(t, commands.WithSources(nil))
	result.ExitCode.Equals(t, 0)
	result.Stderr.Empty(t)
}
