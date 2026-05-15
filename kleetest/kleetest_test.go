package kleetest_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/doron-cohen/klee/kleetest"
	"github.com/stretchr/testify/require"
)

// fakeApp is a minimal Runner for testing kleetest itself.
type fakeApp struct {
	stdout   string
	stderr   string
	exitCode int
}

func (f *fakeApp) Run(_ context.Context, _ []string) int {
	_, _ = fmt.Fprint(os.Stdout, f.stdout)
	_, _ = fmt.Fprint(os.Stderr, f.stderr)
	return f.exitCode
}

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		app      *fakeApp
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "success with stdout",
			app:      &fakeApp{stdout: "hello\n", exitCode: 0},
			wantCode: 0,
			wantOut:  "hello\n",
		},
		{
			name:     "failure with stderr",
			app:      &fakeApp{stderr: "something went wrong\n", exitCode: 1},
			wantCode: 1,
			wantErr:  "something went wrong\n",
		},
		{
			name:     "stdout and stderr independent",
			app:      &fakeApp{stdout: "out\n", stderr: "err\n", exitCode: 2},
			wantCode: 2,
			wantOut:  "out\n",
			wantErr:  "err\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kleetest.Run(t, tt.app)

			result.ExitCode.Equals(t, tt.wantCode)

			if tt.wantOut != "" {
				result.Stdout.Contains(t, tt.wantOut)
			} else {
				result.Stdout.Empty(t)
			}

			if tt.wantErr != "" {
				result.Stderr.Contains(t, tt.wantErr)
			} else {
				result.Stderr.Empty(t)
			}
		})
	}
}

func TestOutputAsserters(t *testing.T) {
	app := &fakeApp{stdout: "exact content\n", exitCode: 0}
	result := kleetest.Run(t, app)

	result.Stdout.Equals(t, "exact content\n")
	result.Stdout.Contains(t, "exact")
	result.Stderr.Empty(t)
	require.Equal(t, "exact content\n", result.Stdout.String())
}
