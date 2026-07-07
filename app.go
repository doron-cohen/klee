package klee

import (
	"context"
	"fmt"
	"os"

	"github.com/doron-cohen/klee/commands"
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
	// SecretStore is used for Secret field resolution during config loading.
	// If nil, the store set via WithSecretStore is used instead.
	SecretStore config.SecretStore
}

// App is a klee application with typed config T.
type App[T any] struct {
	name        string
	version     string
	commands    []*cli.Command
	cfg         *T
	afterLoad   func(cmd *cli.Command) error
	debug       bool
	secretStore config.SecretStore
}

// New creates a new App. Version defaults to version.String(), which
// picks up values injected via ldflags (version.Version, version.Commit,
// version.BuildDate). Use WithVersion to override.
func New[T any](name string, commands []*cli.Command) *App[T] {
	return &App[T]{
		name:     name,
		version:  version.String(),
		commands: commands,
		cfg:      new(T),
	}
}

// WithVersion overrides the version string shown by the built-in version command.
func (a *App[T]) WithVersion(ver string) *App[T] {
	a.version = ver
	return a
}

// WithSecretStore sets the SecretStore used by the secrets CLI command and
// Secret config fields. If the store also implements WritableSecretStore,
// the secrets set subcommand is enabled.
func (a *App[T]) WithSecretStore(s config.SecretStore) *App[T] {
	a.secretStore = s
	return a
}

// LoadConfig loads configuration from files and environment variables.
func (a *App[T]) LoadConfig(opts ConfigOptions[T]) error {
	projectPath := scanConfigFlag(opts.FlagArgs)

	if opts.AfterLoad != nil {
		a.afterLoad = func(cmd *cli.Command) error {
			return opts.AfterLoad(a.cfg, cmd)
		}
	}

	store := opts.SecretStore
	if store == nil {
		store = a.secretStore
	}

	return config.Load(a.cfg, config.Options{
		AppName:     a.name,
		ProjectPath: projectPath,
		DotEnvFiles: opts.DotEnvFiles,
		SecretStore: store,
	})
}

// Run starts the application.
func (a *App[T]) Run(ctx context.Context, args []string) int {
	ctx, cancel := withSignalCancel(ctx)
	defer cancel()

	cmds := a.commands
	if a.secretStore != nil {
		cmds = append([]*cli.Command{commands.SecretsCommand(a.secretStore)}, cmds...)
	}
	cmds = append(cmds, commands.VersionCommand(a.version))
	cmds = append(cmds, commands.ConfigCommand(func(ctx context.Context) any {
		return Config[T](ctx)
	}))

	app := &cli.Command{
		Name:     a.name,
		Flags:    globalFlags,
		Before:   a.before,
		Commands: cmds,
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
