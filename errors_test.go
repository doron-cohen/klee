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

type kindOnly struct {
	kind kleerrors.Kind
	err  error
}

func (e *kindOnly) Error() string             { return e.err.Error() }
func (e *kindOnly) Unwrap() error             { return e.err }
func (e *kindOnly) ErrorKind() kleerrors.Kind { return e.kind }

type hintOnly struct {
	hint string
	err  error
}

func (e *hintOnly) Error() string { return e.err.Error() }
func (e *hintOnly) Unwrap() error { return e.err }
func (e *hintOnly) Hint() string  { return e.hint }

func TestKindAndHintOnDifferentLayers(t *testing.T) {
	base := fmt.Errorf("base")
	tests := []struct {
		name string
		err  error
	}{
		{"hint inside kind", &kindOnly{kleerrors.KindConfig, &hintOnly{"fix the config", base}}},
		{"kind inside hint", &hintOnly{"fix the config", &kindOnly{kleerrors.KindConfig, base}}},
		{"hint inside kind, wrapped", fmt.Errorf("w: %w", &kindOnly{kleerrors.KindConfig, &hintOnly{"fix the config", base}})},
		{"kind inside hint, wrapped", fmt.Errorf("w: %w", &hintOnly{"fix the config", &kindOnly{kleerrors.KindConfig, base}})},
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
