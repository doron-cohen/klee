package testapp_test

import (
	"testing"
)

// A mistyped command and a flag that does not exist are both the user's
// mistake, so both exit 2. Left to urfave they exit 3 and 1 respectively — 3
// because an unrecognised name is handled at the root before Run returns, and
// 1 because a parse error carries no kind.
func TestUsageErrorsExitTwo(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "unknown top-level command",
			args: []string{"frobnicate"},
			want: `unknown command "frobnicate"`,
		},
		{
			name: "unknown subcommand",
			args: []string{"msg", "frobnicate"},
			want: `unknown command "frobnicate"`,
		},
		{
			name: "unknown subcommand of a built-in",
			args: []string{"config", "frobnicate"},
			want: `unknown command "frobnicate"`,
		},
		{
			name: "undefined flag",
			args: []string{"--nosuchflag"},
			want: "flag provided but not defined",
		},
		{
			name: "undefined flag after a command",
			args: []string{"echo", "--nosuchflag"},
			want: "flag provided but not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(t, tt.args...)
			result.ExitCode.Equals(t, 2)
			result.Stderr.Contains(t, tt.want)
		})
	}
}

// Classifying usage errors must not turn working commands into failures.
func TestValidInvocationsUnaffected(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"a command", []string{"echo"}},
		{"a subcommand", []string{"msg", "success"}},
		{"a built-in", []string{"version"}},
		{"help", []string{"--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt.args...).ExitCode.Equals(t, 0)
		})
	}
}

// A command's own error keeps its kind; only unclassified usage becomes a user
// error. Without this, the fix would flatten every exit code to 2.
func TestCommandErrorsKeepTheirKind(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"user error", []string{"fail", "user"}, 2},
		{"internal error", []string{"fail", "internal"}, 3},
		{"error with no kind", []string{"fail", "plain"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt.args...).ExitCode.Equals(t, tt.want)
		})
	}
}
