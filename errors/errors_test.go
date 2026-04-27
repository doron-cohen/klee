package errors_test

import (
	"testing"

	"github.com/doron-cohen/klee/errors"
	"github.com/stretchr/testify/require"
)

// testError is a concrete error type used to verify the interfaces.
type testError struct {
	kind errors.Kind
	msg  string
	hint string
}

func (e *testError) Error() string        { return e.msg }
func (e *testError) ErrorKind() errors.Kind { return e.kind }
func (e *testError) Hint() string         { return e.hint }

func TestKindValues(t *testing.T) {
	require.Equal(t, errors.Kind(1), errors.KindUser)
	require.Equal(t, errors.Kind(2), errors.KindInternal)
	require.Equal(t, errors.Kind(3), errors.KindConfig)
	require.NotEqual(t, errors.KindUser, errors.KindInternal)
	require.NotEqual(t, errors.KindInternal, errors.KindConfig)
}

func TestKinderInterface(t *testing.T) {
	tests := []struct {
		name string
		err  *testError
		kind errors.Kind
	}{
		{"user", &testError{kind: errors.KindUser, msg: "bad input"}, errors.KindUser},
		{"internal", &testError{kind: errors.KindInternal, msg: "bug"}, errors.KindInternal},
		{"config", &testError{kind: errors.KindConfig, msg: "bad config"}, errors.KindConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var k errors.Kinder = tt.err
			require.Equal(t, tt.kind, k.ErrorKind())
		})
	}
}

func TestHinterInterface(t *testing.T) {
	tests := []struct {
		name string
		err  *testError
		hint string
	}{
		{"with hint", &testError{hint: "try this instead"}, "try this instead"},
		{"empty hint", &testError{hint: ""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h errors.Hinter = tt.err
			require.Equal(t, tt.hint, h.Hint())
		})
	}
}
