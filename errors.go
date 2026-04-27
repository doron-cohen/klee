package klee

import (
	"fmt"

	kleerrors "github.com/doron-cohen/klee/errors"
)

func exitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	if k, ok := err.(kleerrors.Kinder); ok {
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

func renderError(err error, debug bool) string {
	msg := err.Error()

	if h, ok := err.(kleerrors.Hinter); ok {
		if hint := h.Hint(); hint != "" {
			msg += "\nHint: " + hint
		}
	}

	if k, ok := err.(kleerrors.Kinder); ok {
		if k.ErrorKind() == kleerrors.KindInternal && !debug {
			msg += "\nRun with --debug for more details."
		}
	}

	return fmt.Sprintf("Error: %s", msg)
}
