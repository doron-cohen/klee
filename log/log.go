package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	internallog "github.com/doron-cohen/klee/internal/log"
	"github.com/doron-cohen/klee/xdg"
)

// Provider is implemented by any config struct that embeds log.Config.
// App[T] checks this interface to auto-wire logging.
type Provider interface {
	LogConfig() Config
}

// SetupOptions controls log setup behaviour.
type SetupOptions struct {
	// AppName is used to derive the default file path via XDG.
	AppName string
	// NoColor disables ANSI color sequences in the pretty handler.
	NoColor bool
}

type contextKey struct{}

// Setup builds slog handlers from cfg, sets slog.Default, stores the logger
// in ctx, and returns both the updated context and the logger.
func Setup(ctx context.Context, cfg Config, opts SetupOptions) (context.Context, *slog.Logger, error) {
	var handlers []slog.Handler

	if cfg.Console.Enabled {
		level, err := parseLevel(cfg.Console.Level)
		if err != nil {
			return ctx, nil, fmt.Errorf("log: console level: %w", err)
		}
		h := internallog.NewConsoleHandler(os.Stderr, level, cfg.Console.Format, opts.NoColor)
		handlers = append(handlers, h)
	}

	if cfg.File.Enabled {
		level, err := parseLevel(cfg.File.Level)
		if err != nil {
			return ctx, nil, fmt.Errorf("log: file level: %w", err)
		}
		path := cfg.File.Path
		if path == "" {
			path = filepath.Join(xdg.New(opts.AppName).DataHome(), opts.AppName+".log")
		}
		h := internallog.NewFileHandler(internallog.FileOptions{
			Path:       path,
			Level:      level,
			Format:     cfg.File.Format,
			MaxSizeMB:  cfg.File.Rotation.MaxSizeMB,
			MaxBackups: cfg.File.Rotation.MaxBackups,
			MaxAgeDays: cfg.File.Rotation.MaxAgeDays,
		})
		handlers = append(handlers, h)
	}

	logger := buildLogger(handlers)
	slog.SetDefault(logger)
	ctx = WithCtx(ctx, logger)
	return ctx, logger, nil
}

// FromCtx returns the *slog.Logger stored in ctx.
// Falls back to slog.Default() if none is stored.
func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// WithCtx stores logger in ctx and returns the updated context.
func WithCtx(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func buildLogger(handlers []slog.Handler) *slog.Logger {
	switch len(handlers) {
	case 0:
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	case 1:
		return slog.New(handlers[0])
	default:
		return slog.New(internallog.NewMultiHandler(handlers))
	}
}

func parseLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q (valid: debug, info, warn, error)", s)
	}
}
