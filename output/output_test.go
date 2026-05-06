package output_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/doron-cohen/klee"
	"github.com/doron-cohen/klee/output"
	"github.com/stretchr/testify/require"
)

func newOutput(t *testing.T, buf *bytes.Buffer, flags klee.RunFlags) *output.Output {
	t.Helper()
	return output.New(buf, flags)
}

func TestSuccess(t *testing.T) {
	var buf bytes.Buffer
	out := newOutput(t, &buf, klee.RunFlags{NoColor: true})
	out.Success("everything worked")
	require.Contains(t, buf.String(), "everything worked")
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	out := newOutput(t, &buf, klee.RunFlags{NoColor: true})
	out.Warn("something is off")
	require.Contains(t, buf.String(), "something is off")
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	out := newOutput(t, &buf, klee.RunFlags{NoColor: true})
	out.Error("something failed")
	require.Contains(t, buf.String(), "something failed")
}

func TestHint(t *testing.T) {
	var buf bytes.Buffer
	out := newOutput(t, &buf, klee.RunFlags{NoColor: true})
	out.Hint("try this instead")
	require.Contains(t, buf.String(), "try this instead")
}

func TestQuietSuppressesWarnAndHint(t *testing.T) {
	tests := []struct {
		name   string
		write  func(*output.Output)
		silent bool
	}{
		{"warn suppressed", func(o *output.Output) { o.Warn("warning") }, true},
		{"hint suppressed", func(o *output.Output) { o.Hint("hint") }, true},
		{"success shown", func(o *output.Output) { o.Success("ok") }, false},
		{"error shown", func(o *output.Output) { o.Error("err") }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			out := newOutput(t, &buf, klee.RunFlags{NoColor: true, Quiet: true})
			tt.write(out)
			if tt.silent {
				require.Empty(t, buf.String())
			} else {
				require.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestJSONSuppressesAll(t *testing.T) {
	for _, name := range []string{"success", "warn", "error", "hint"} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			out := newOutput(t, &buf, klee.RunFlags{NoColor: true, JSON: true})
			switch name {
			case "success":
				out.Success("msg")
			case "warn":
				out.Warn("msg")
			case "error":
				out.Error("msg")
			case "hint":
				out.Hint("msg")
			}
			require.Empty(t, buf.String())
		})
	}
}

func TestNoColorSkipsANSI(t *testing.T) {
	var buf bytes.Buffer
	out := newOutput(t, &buf, klee.RunFlags{NoColor: true})
	out.Success("plain text")
	require.NotContains(t, buf.String(), "\x1b[")
}

func TestFromCtxEmptyContext(t *testing.T) {
	// Should not panic on empty context — returns usable Output
	out := output.FromCtx(context.Background())
	require.NotNil(t, out)
	// Writing should not panic (goes to os.Stderr which is discarded in tests)
	out.Success("test") // no assertion, just checking no panic
}
