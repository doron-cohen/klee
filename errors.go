package klee

import (
	"errors"
	"fmt"

	kleerrors "github.com/doron-cohen/klee/errors"
)

func exitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	var k kleerrors.Kinder
	if errors.As(err, &k) {
		switch k.ErrorKind() {
		case kleerrors.KindUser:
			return 2
		case kleerrors.KindInternal:
			return 3
		case kleerrors.KindConfig:
			return 4
		}
	}
	return 1
}

// exitCode resolves the exit code for err: matchers first (in registration
// order), then the kind map from WithExitCodes, then the built-in defaults.
func (a *App[T]) exitCode(err error) int {
	if err == nil {
		return 0
	}
	for _, m := range a.exitMatchers {
		if code, ok := m(err); ok {
			return code
		}
	}
	if len(a.exitCodes) > 0 {
		var k kleerrors.Kinder
		if errors.As(err, &k) {
			if code, ok := a.exitCodes[k.ErrorKind()]; ok {
				return code
			}
		}
	}
	return exitCodeForError(err)
}

func renderError(err error, debug bool) string {
	msg := err.Error()

	var h kleerrors.Hinter
	if errors.As(err, &h) {
		if hint := h.Hint(); hint != "" {
			msg += "\nHint: " + hint
		}
	}

	var k kleerrors.Kinder
	if errors.As(err, &k) {
		if k.ErrorKind() == kleerrors.KindInternal && !debug {
			msg += "\nRun with --debug for more details."
		}
	}

	return fmt.Sprintf("Error: %s", msg)
}
