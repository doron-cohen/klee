// Package kleetest provides a test harness for klee applications.
package kleetest

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
)

// Result holds the captured output of a Run call.
type Result struct {
	ExitCode *exitCodeAsserter
	Stdout   *outputAsserter
	Stderr   *outputAsserter
}

// Runner is anything that can run with args and return an exit code.
// *klee.App[T] satisfies this interface.
type Runner interface {
	Run(ctx context.Context, args []string) int
}

// Run executes the app with the given args and captures stdout, stderr, and exit code.
func Run(t *testing.T, app Runner, args ...string) *Result {
	t.Helper()

	var code int
	stdout, stderr := capture(t, func() {
		code = app.Run(context.Background(), append([]string{"app"}, args...))
	})

	return &Result{
		ExitCode: &exitCodeAsserter{value: code},
		Stdout:   &outputAsserter{content: stdout},
		Stderr:   &outputAsserter{content: stderr},
	}
}

// capture redirects os.Stdout and os.Stderr, runs fn, and returns the captured output.
func capture(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()

	origStdout, origStderr := os.Stdout, os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("kleetest: failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("kleetest: failed to create stderr pipe: %v", err)
	}

	os.Stdout, os.Stderr = wOut, wErr

	fn()

	if err := wOut.Close(); err != nil {
		t.Logf("kleetest: closing stdout pipe writer: %v", err)
	}
	if err := wErr.Close(); err != nil {
		t.Logf("kleetest: closing stderr pipe writer: %v", err)
	}
	os.Stdout, os.Stderr = origStdout, origStderr

	var bufOut, bufErr bytes.Buffer
	if _, err := io.Copy(&bufOut, rOut); err != nil {
		t.Fatalf("kleetest: reading stdout: %v", err)
	}
	if _, err := io.Copy(&bufErr, rErr); err != nil {
		t.Fatalf("kleetest: reading stderr: %v", err)
	}
	if err := rOut.Close(); err != nil {
		t.Logf("kleetest: closing stdout pipe reader: %v", err)
	}
	if err := rErr.Close(); err != nil {
		t.Logf("kleetest: closing stderr pipe reader: %v", err)
	}

	return bufOut.String(), bufErr.String()
}
