package version

import (
	"bytes"
	"text/template"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Template is the Go template used by String().
// Override to change the version format.
var Template = "{{.Version}} ({{.Commit}}, {{.BuildDate}})"

// String formats Version, Commit, and BuildDate using Template.
func String() string {
	t, err := template.New("version").Parse(Template)
	if err != nil {
		return Version
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, map[string]string{
		"Version":   Version,
		"Commit":    Commit,
		"BuildDate": BuildDate,
	}); err != nil {
		return Version
	}
	return buf.String()
}
