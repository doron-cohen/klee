package klee

import (
	"context"
	"fmt"

	kleerrors "github.com/doron-cohen/klee/errors"
	"github.com/urfave/cli/v3"
)

// usageError is a mistake in what was typed: a command that does not exist, or
// a flag the parser does not have.
//
// Both reach the shell as the wrong exit code unless they are caught. urfave
// answers an unrecognised name with Exit(msg, 3) and handles it at the root
// before Run returns, so the process exits 3 — its number for "no help topic",
// indistinguishable from a claim that the program broke. A flag parse error
// does return, but carries no Kind, so it maps to 1 alongside genuine runtime
// failures. Both are the user's mistake and exit 2.
type usageError struct{ msg string }

func (e *usageError) Error() string             { return e.msg }
func (e *usageError) ErrorKind() kleerrors.Kind { return kleerrors.KindUser }

func unknownCommand(name string) error {
	return &usageError{msg: fmt.Sprintf("unknown command %q", name)}
}

// classifyUsage gives every command in the tree an answer for those two
// mistakes, so a wrong name is caught at whatever depth it was typed.
//
// The two hooks cannot work the same way: OnUsageError returns an error, which
// flows back through Run, while CommandNotFound returns nothing at all. The
// name is therefore reported through found and raised by the caller once Run
// has finished.
func classifyUsage(cmd *cli.Command, found func(string)) {
	cmd.CommandNotFound = func(_ context.Context, _ *cli.Command, name string) {
		found(name)
	}
	cmd.OnUsageError = func(_ context.Context, _ *cli.Command, err error, _ bool) error {
		return &usageError{msg: err.Error()}
	}

	for _, sub := range cmd.Commands {
		classifyUsage(sub, found)
	}
}
