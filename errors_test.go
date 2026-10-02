package klee

import (
	"fmt"
	"testing"

	kleerrors "github.com/doron-cohen/klee/errors"
	"github.com/stretchr/testify/require"
)

type kindErr struct {
	kind kleerrors.Kind
	hint string
}

func (e *kindErr) Error() string             { return "boom" }
func (e *kindErr) ErrorKind() kleerrors.Kind { return e.kind }
func (e *kindErr) Hint() string              { return e.hint }

func TestErrorKindSurvivesWrapping(t *testing.T) {
	inner := &kindErr{kind: kleerrors.KindConfig, hint: "fix the config"}
	tests := []struct {
		name string
		err  error
	}{
		{"bare", inner},
		{"wrapped", fmt.Errorf("config validation failed: %w", inner)},
		{"wrapped twice", fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", inner))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, 4, exitCodeForError(tt.err))
			require.Contains(t, renderError(tt.err, false), "\nHint: fix the config")
		})
	}
}

func TestInternalKindWrappedAddsDebugHint(t *testing.T) {
	err := fmt.Errorf("wrap: %w", &kindErr{kind: kleerrors.KindInternal})
	require.Equal(t, 3, exitCodeForError(err))
	require.Contains(t, renderError(err, false), "Run with --debug")
	require.NotContains(t, renderError(err, true), "Run with --debug")
}
