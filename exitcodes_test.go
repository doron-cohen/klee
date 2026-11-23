package klee_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/doron-cohen/klee"
	kleerrors "github.com/doron-cohen/klee/errors"
	"github.com/doron-cohen/klee/kleetest"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

var (
	kindConflict = kleerrors.CustomKind(0)
	kindNotFound = kleerrors.CustomKind(1)
)

type appErr struct {
	kind kleerrors.Kind
	hint string
}

func (e *appErr) Error() string             { return "boom" }
func (e *appErr) ErrorKind() kleerrors.Kind { return e.kind }
func (e *appErr) Hint() string              { return e.hint }

var errSentinel = errors.New("sentinel")

type empty struct{}

func newApp(fail error) *klee.App[empty] {
	return klee.New[empty]("app", []*cli.Command{{
		Name:   "fail",
		Action: func(context.Context, *cli.Command) error { return fail },
	}})
}

func TestCustomKindsDistinctFromBuiltins(t *testing.T) {
	for _, b := range []kleerrors.Kind{kleerrors.KindUser, kleerrors.KindInternal, kleerrors.KindConfig} {
		require.NotEqual(t, b, kindConflict)
		require.NotEqual(t, b, kindNotFound)
	}
	require.NotEqual(t, kindConflict, kindNotFound)
}

func TestWithExitCodesCustomAndOverride(t *testing.T) {
	codes := map[kleerrors.Kind]int{kindConflict: 3, kindNotFound: 4, kleerrors.KindUser: 2, kleerrors.KindConfig: 9}
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"custom", &appErr{kind: kindConflict}, 3},
		{"custom wrapped", fmt.Errorf("w: %w", &appErr{kind: kindNotFound}), 4},
		{"override builtin", &appErr{kind: kleerrors.KindConfig}, 9},
		{"unmapped builtin keeps default", &appErr{kind: kleerrors.KindInternal}, 3},
		{"untyped", errors.New("x"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := kleetest.Run(t, newApp(tt.err).WithExitCodes(codes), "fail")
			r.ExitCode.Equals(t, tt.want)
		})
	}
}

func TestCustomKindHintRendering(t *testing.T) {
	app := newApp(fmt.Errorf("ctx: %w", &appErr{kind: kindConflict, hint: "retry later"})).
		WithExitCodes(map[kleerrors.Kind]int{kindConflict: 3})
	r := kleetest.Run(t, app, "fail")
	r.ExitCode.Equals(t, 3)
	r.Stderr.Contains(t, "Hint: retry later")
	require.NotContains(t, r.Stderr.String(), "--debug")
}

func TestWithExitMatcherPrecedence(t *testing.T) {
	app := newApp(fmt.Errorf("w: %w", errSentinel)).
		WithExitCodes(map[kleerrors.Kind]int{kindConflict: 3}).
		WithExitMatcher(func(err error) (int, bool) { return 0, false }).
		WithExitMatcher(func(err error) (int, bool) { return 6, errors.Is(err, errSentinel) })
	kleetest.Run(t, app, "fail").ExitCode.Equals(t, 6)

	// matcher wins over the kind map
	app2 := newApp(&appErr{kind: kindConflict}).
		WithExitCodes(map[kleerrors.Kind]int{kindConflict: 3}).
		WithExitMatcher(func(err error) (int, bool) { return 7, true })
	kleetest.Run(t, app2, "fail").ExitCode.Equals(t, 7)
}

func TestUnconfiguredAppKeepsDefaults(t *testing.T) {
	kleetest.Run(t, newApp(&appErr{kind: kleerrors.KindUser}), "fail").ExitCode.Equals(t, 2)
	kleetest.Run(t, newApp(&appErr{kind: kindConflict}), "fail").ExitCode.Equals(t, 1)
}
