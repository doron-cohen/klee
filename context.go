package klee

import "context"

type contextKey int

const (
	configKey   contextKey = iota
	runFlagsKey contextKey = iota
)

// RunFlags holds the values of global run-control flags.
type RunFlags struct {
	Debug   bool
	Quiet   bool
	JSON    bool
	NoColor bool
}

// Config returns the typed config stored in ctx.
func Config[T any](ctx context.Context) *T {
	v, _ := ctx.Value(configKey).(*T)
	return v
}

// GetRunFlags returns the RunFlags stored in ctx.
func GetRunFlags(ctx context.Context) RunFlags {
	v, _ := ctx.Value(runFlagsKey).(RunFlags)
	return v
}

func withConfig[T any](ctx context.Context, cfg *T) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

func withRunFlags(ctx context.Context, flags RunFlags) context.Context {
	return context.WithValue(ctx, runFlagsKey, flags)
}
