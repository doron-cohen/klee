package klee

import (
	"context"
	"fmt"
	"os"

	"github.com/doron-cohen/klee/config"
	kleelog "github.com/doron-cohen/klee/log"
	"github.com/doron-cohen/klee/version"
	"github.com/urfave/cli/v3"
)

// ConfigOptions controls config loading.
type ConfigOptions[T any] struct {
	// FlagArgs is typically os.Args — scanned for --config before full parse.
	FlagArgs []string
	// AfterLoad is called after config is loaded, with typed config and full CLI flag access.
	AfterLoad func(cfg *T, cmd *cli.Command) error
	// DotEnvFiles are .env files to load KEY=VALUE pairs from.
	// Real environment variables take precedence over values in these files.
	DotEnvFiles []string
}

// App is a klee application with typed config T.
type App[T any] struct {
	name      string
	version   string
	commands  []*cli.Command
	cfg       *T
	afterLoad func(cmd *cli.Command) error
	debug     bool
}

// New creates a new App.
func New[T any](name, ver string, commands []*cli.Command) *App[T] {
	return &App[T]{
		name:     name,
		version:  ver,
		commands: commands,
		cfg:      new(T),
	}
}

// LoadConfig loads configuration from files and environment variables.
func (a *App[T]) LoadConfig(opts ConfigOptions[T]) error {
	projectPath := scanConfigFlag(opts.FlagArgs)

	if opts.AfterLoad != nil {
		a.afterLoad = func(cmd *cli.Command) error {
			return opts.AfterLoad(a.cfg, cmd)
		}
	}

	return config.Load(a.cfg, config.Options{
		AppName:     a.name,
		ProjectPath: projectPath,
		DotEnvFiles: opts.DotEnvFiles,
	})
}

// Run starts the application.
func (a *App[T]) Run(ctx context.Context, args []string) int {
	ctx, cancel := withSignalCancel(ctx)
	defer cancel()

	app := &cli.Command{
		Name:     a.name,
		Flags:    globalFlags,
		Before:   a.before,
		Commands: append(a.commands, versionCommand(a.version)),
	}

	if err := app.Run(ctx, args); err != nil {
		fmt.Fprintln(os.Stderr, renderError(err, a.debug))
		return exitCodeForError(err)
	}

	return 0
}

func (a *App[T]) before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	a.debug = cmd.Bool("debug")

	ctx = withRunFlags(ctx, RunFlags{
		Debug:   a.debug,
		Quiet:   cmd.Bool("quiet"),
		JSON:    cmd.Bool("json"),
		NoColor: cmd.Bool("no-color"),
	})
	ctx = withConfig(ctx, a.cfg)

	if provider, ok := any(a.cfg).(kleelog.Provider); ok {
		logCfg := provider.LogConfig()
		if level := cmd.String("log-level"); level != "" {
			logCfg.Console.Level = level
			logCfg.File.Level = level
		}
		if format := cmd.String("log-format"); format != "" {
			logCfg.Console.Format = format
		}
		var err error
		ctx, _, err = kleelog.Setup(ctx, logCfg, kleelog.SetupOptions{
			AppName: a.name,
			NoColor: cmd.Bool("no-color"),
		})
		if err != nil {
			return ctx, err
		}
	}

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
				fmt.Println(version.Version)
			} else {
				fmt.Println(ver)
			}
			return nil
		},
	}
}
