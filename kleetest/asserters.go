package kleetest

import (
	"strings"
	"testing"
)

type exitCodeAsserter struct {
	value int
}

func (a *exitCodeAsserter) Equals(t *testing.T, expected int) {
	t.Helper()
	if a.value != expected {
		t.Errorf("exit code: got %d, want %d", a.value, expected)
	}
}

type outputAsserter struct {
	content string
}

func (a *outputAsserter) Contains(t *testing.T, substr string) {
	t.Helper()
	if !strings.Contains(a.content, substr) {
		t.Errorf("output does not contain %q\ngot: %q", substr, a.content)
	}
}

func (a *outputAsserter) Equals(t *testing.T, expected string) {
	t.Helper()
	if a.content != expected {
		t.Errorf("output mismatch\ngot:  %q\nwant: %q", a.content, expected)
	}
}

func (a *outputAsserter) Empty(t *testing.T) {
	t.Helper()
	if a.content != "" {
		t.Errorf("expected empty output, got: %q", a.content)
	}
}

// String returns the raw captured content, useful for custom assertions.
func (a *outputAsserter) String() string {
	return a.content
}
