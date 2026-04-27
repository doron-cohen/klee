package version_test

import (
	"testing"

	"github.com/doron-cohen/klee/version"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildDate string
		template  string
		expected  string
	}{
		{
			name:      "default template",
			version:   "1.2.3",
			commit:    "abc1234",
			buildDate: "2026-04-27",
			expected:  "1.2.3 (abc1234, 2026-04-27)",
		},
		{
			name:     "version only template",
			version:  "2.0.0",
			template: "{{.Version}}",
			expected: "2.0.0",
		},
		{
			name:      "custom template",
			version:   "3.0.0",
			commit:    "xyz",
			buildDate: "2026-01-01",
			template:  "v{{.Version}}-{{.Commit}}",
			expected:  "v3.0.0-xyz",
		},
		{
			name:     "invalid template falls back to Version",
			version:  "1.0.0",
			template: "{{.Invalid",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := version.Version
			origCommit := version.Commit
			origBuildDate := version.BuildDate
			origTemplate := version.Template
			t.Cleanup(func() {
				version.Version = orig
				version.Commit = origCommit
				version.BuildDate = origBuildDate
				version.Template = origTemplate
			})

			version.Version = tt.version
			version.Commit = tt.commit
			version.BuildDate = tt.buildDate
			if tt.template != "" {
				version.Template = tt.template
			}

			require.Equal(t, tt.expected, version.String())
		})
	}
}
