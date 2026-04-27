package klee

import (
	"context"
	"fmt"
	"os"

	"github.com/doron-cohen/klee/config"
	"github.com/urfave/cli/v3"
)

// ConfigOptions controls config loading.
type ConfigOptions struct {
	// FlagArgs is typically os.Args — scanned for --config before full parse.
	FlagArgs []string
	// AfterLoad is called after config is loaded, with full CLI flag access.
	AfterLoad func(cmd *cli.Command) error
}

// App is a klee application with typed config T.
type App[T any] struct {
	name      string
	version   string
	commands  []*cli.Command
	cfg       *T
	afterLoad func(cmd *cli.Command) error
}

// New creates a new App.
func New[T any](name, version string, commands []*cli.Command) *App[T] {
	return &App[T]{
		name:     name,
		version:  version,
		commands: commands,
		cfg:      new(T),
	}
}

// LoadConfig loads configuration from files and environment variables.
func (a *App[T]) LoadConfig(opts ConfigOptions) error {
	projectPath := scanConfigFlag(opts.FlagArgs)
	a.afterLoad = opts.AfterLoad

	return config.Load(a.cfg, config.Options{
		AppName:     a.name,
		ProjectPath: projectPath,
	})
}

// Run starts the application.
func (a *App[T]) Run(ctx context.Context, args []string) int {
	ctx, cancel := withSignalCancel(ctx)
	defer cancel()

	app := &cli.Command{
		Name:    a.name,
		Version: a.version,
		Flags:   globalFlags,
		Before: a.before,
		Commands: append(a.commands, versionCommand(a.version)),
	}

	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, renderError(err, false))
		return exitCodeForError(err)
	}

	return 0
}

func (a *App[T]) before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	flags := RunFlags{
		Debug:   cmd.Bool("debug"),
		Quiet:   cmd.Bool("quiet"),
		JSON:    cmd.Bool("json"),
		NoColor: cmd.Bool("no-color"),
	}

	ctx = withRunFlags(ctx, flags)
	ctx = withConfig(ctx, a.cfg)

	if a.afterLoad != nil {
		return ctx, a.afterLoad(cmd)
	}
	return ctx, nil
}

// scanConfigFlag does a simple scan of args for --config <path>.
func scanConfigFlag(args []string) string {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func versionCommand(ver string) *cli.Command {
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
				fmt.Println(ver)
			} else {
				fmt.Println(ver)
			}
			return nil
		},
	}
}
