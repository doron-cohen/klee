package klee

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// withSignalCancel returns a context that is cancelled on SIGTERM or SIGINT.
func withSignalCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(ch)
	}()

	return ctx, cancel
}
