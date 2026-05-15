package testapp

import (
	"context"
	"fmt"
	"os"

	"github.com/doron-cohen/klee"
	kleerrors "github.com/doron-cohen/klee/errors"
	kleelog "github.com/doron-cohen/klee/log"
	"github.com/doron-cohen/klee/output"
	"github.com/urfave/cli/v3"
)

// echoCmd prints config values to stdout.
var echoCmd = &cli.Command{
	Name:  "echo",
	Usage: "print config values",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		cfg := klee.Config[Config](ctx)
		_, _ = fmt.Fprintf(os.Stdout, "host=%s port=%d\n", cfg.Host, cfg.Port)
		return nil
	},
}

// flagsCmd prints run flag values to stdout.
var flagsCmd = &cli.Command{
	Name:  "flags",
	Usage: "print run flag values",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		rf := klee.GetRunFlags(ctx)
		_, _ = fmt.Fprintf(os.Stdout, "debug=%v quiet=%v json=%v no-color=%v\n",
				rf.Debug, rf.Quiet, rf.JSON, rf.NoColor)
		return nil
	},
}

// logCmd logs a message at info level using the context logger.
var logCmd = &cli.Command{
	Name:  "log",
	Usage: "log a message via the context logger",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		kleelog.FromCtx(ctx).Info("log command ran", "key", "value")
		return nil
	},
}

// msgCmd exercises all output message types.
var msgCmd = &cli.Command{
	Name:  "msg",
	Usage: "print a message via output package",
	Commands: []*cli.Command{
		{
			Name: "success",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				output.FromCtx(ctx).Success("success message")
				return nil
			},
		},
		{
			Name: "warn",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				output.FromCtx(ctx).Warn("warn message")
				return nil
			},
		},
		{
			Name: "error",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				output.FromCtx(ctx).Error("error message")
				return nil
			},
		},
		{
			Name: "hint",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				output.FromCtx(ctx).Hint("hint message")
				return nil
			},
		},
	},
}

// failCmd groups commands that return different error types.
var failCmd = &cli.Command{
	Name:  "fail",
	Usage: "commands that return errors (for testing)",
	Commands: []*cli.Command{
		{
			Name: "user",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return &testUserError{msg: "bad input", hint: "check your input and try again"}
			},
		},
		{
			Name: "internal",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return &testInternalError{msg: "unexpected failure"}
			},
		},
		{
			Name: "plain",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return fmt.Errorf("plain error")
			},
		},
	},
}

type testUserError struct {
	msg  string
	hint string
}

func (e *testUserError) Error() string             { return e.msg }
func (e *testUserError) ErrorKind() kleerrors.Kind { return kleerrors.KindUser }
func (e *testUserError) Hint() string              { return e.hint }

type testInternalError struct {
	msg string
}

func (e *testInternalError) Error() string             { return e.msg }
func (e *testInternalError) ErrorKind() kleerrors.Kind { return kleerrors.KindInternal }
