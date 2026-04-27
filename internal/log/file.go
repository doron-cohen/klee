package log

import (
	"log/slog"

	"gopkg.in/lumberjack.v2"
)

// FileOptions configures the file sink handler.
type FileOptions struct {
	Path       string
	Level      slog.Level
	Format     string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
}

// NewFileHandler returns a slog.Handler writing to a rotating log file.
func NewFileHandler(opts FileOptions) slog.Handler {
	lj := &lumberjack.Logger{
		Filename:   opts.Path,
		MaxSize:    opts.MaxSizeMB,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAgeDays,
		Compress:   true,
	}
	hopts := &slog.HandlerOptions{Level: opts.Level}
	switch opts.Format {
	case "text":
		return slog.NewTextHandler(lj, hopts)
	default: // "json"
		return slog.NewJSONHandler(lj, hopts)
	}
}
