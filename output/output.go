package output

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/doron-cohen/klee"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))  // red
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // dim gray
)

// Output writes styled messages to a writer (typically stderr).
type Output struct {
	w       io.Writer
	noColor bool
	json    bool
	quiet   bool
}

// New creates an Output with explicit writer and flags.
func New(w io.Writer, flags klee.RunFlags) *Output {
	return &Output{
		w:       w,
		noColor: flags.NoColor,
		json:    flags.JSON,
		quiet:   flags.Quiet,
	}
}

// FromCtx derives an Output from RunFlags already in ctx.
// Uses os.Stderr. Zero cost — no setup step required.
func FromCtx(ctx context.Context) *Output {
	flags := klee.GetRunFlags(ctx)
	return &Output{
		w:       os.Stderr,
		noColor: flags.NoColor || !isTTY(os.Stderr) || isNoColor(),
		json:    flags.JSON,
		quiet:   flags.Quiet,
	}
}

// Success prints a success message. Suppressed in --json mode.
func (o *Output) Success(format string, args ...any) {
	if o.json {
		return
	}
	o.print(successStyle, format, args...)
}

// Warn prints a warning message. Suppressed in --quiet and --json modes.
func (o *Output) Warn(format string, args ...any) {
	if o.quiet || o.json {
		return
	}
	o.print(warnStyle, format, args...)
}

// Error prints an error message. Suppressed in --json mode only.
func (o *Output) Error(format string, args ...any) {
	if o.json {
		return
	}
	o.print(errorStyle, format, args...)
}

// Hint prints a hint message. Suppressed in --quiet and --json modes.
func (o *Output) Hint(format string, args ...any) {
	if o.quiet || o.json {
		return
	}
	o.print(hintStyle, format, args...)
}

func (o *Output) print(style lipgloss.Style, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if o.noColor {
		_, _ = fmt.Fprintln(o.w, msg)
		return
	}
	_, _ = fmt.Fprintln(o.w, style.Render(msg))
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

func isNoColor() bool {
	_, set := os.LookupEnv("NO_COLOR")
	return set
}
