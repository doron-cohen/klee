package log

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// NewConsoleHandler returns a slog.Handler for the console sink.
// format: "pretty" or "json". noColor disables ANSI even on TTYs.
func NewConsoleHandler(w io.Writer, level slog.Level, format string, noColor bool) slog.Handler {
	switch format {
	case "json":
		return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	default: // "pretty"
		return tint.NewHandler(w, &tint.Options{
			Level:   level,
			NoColor: noColor || !isTTY(w),
		})
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
