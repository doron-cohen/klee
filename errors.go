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
