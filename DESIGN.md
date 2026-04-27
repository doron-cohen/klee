# klee design document

`github.com/doron-cohen/klee`

## Guiding principle

**Same defaults while allowing flexibility.**
Every module ships with sensible defaults. Everything can be overridden or disabled.

---

## `version`

Global vars set via ldflags:
```
version.Version, version.Commit, version.BuildDate
```
- `version.String()` — formats vars using default Go template
- `version.Template` — override to change format
- Built-in `version` command, `--short` flag for version number only
- Configurable: override vars, custom template, disable the command

---

## `errors`

Interfaces only, no concrete types:
```
Kinder — ErrorKind() Kind
Hinter — Hint() string
```
Kind constants: `KindUser`, `KindInternal`, `KindConfig`

Each package owns its error types and implements these interfaces.

CLI owns exit codes:
```
1 = uncategorized (fallback)
2 = KindUser
3 = KindInternal
4 = KindConfig
```
Extensible via registered matchers.

---

## `config`

Struct-driven, tags define behavior:
```
yaml:    — config file key
env:     — environment variable name
default: — fallback value
secret:  — redacted in config print
```
Doc strings via Go field comments, extracted via AST for doc generation.

Precedence: flags → env vars → project file → user file → system file → defaults

XDG paths for user/system files, driven by app name.

Validation opt-in:
- implement `Validate() error` on config struct
- or pass a validator option for struct tag validation

Built-in commands: `config validate`, `config print`

App composes package configs via embedding:
```go
type AppConfig struct {
    log.Config `yaml:"log"`
    DB         db.Config `yaml:"db"`
}
```

---

## `log`

Uses slog directly — no wrapper.
klee wires level, format, and output from `log.Config`.

Two independent sinks:

Console:
- enabled by default
- level: info by default
- format: pretty (human) or json

File:
- disabled by default
- path: XDG data dir by default
- level: debug by default
- format: json by default
- rotation: max size, max backups, max age

Each sink configured independently.
Global flags wired automatically: `--log-level`, `--log-format`

---

## `output`

Wires Charm/lipgloss — not built from scratch.

TTY detection, `NO_COLOR`, `FORCE_COLOR` respected.
Styling disabled when not a TTY or `--json`.

Color palette:
```
green  — success
yellow — warning
red    — error
dim    — secondary info
cyan   — paths, names, links
```

- Styled message functions: Success, Warn, Error, Hint
- Table
- Key-value (single object display)
- Confirm prompt — skipped under `--no-input`, fails in non-TTY
- Spinner — stderr only, hidden when not TTY
- Terminal width aware
- `--json`: stable, no styling, structured to stdout
- `--quiet`: suppresses non-essential output
- Context-propagated

Future: pager, terminal links, table action decorator pattern

---

## `cli`

Single entry point, app name drives XDG paths and built-in command names.
Config type is generic — typed access inside commands, no casting.

Config loading:
- pre-scans args for `--config` before full parse
- `flag:` tag on config fields declares which flag overrides which field
- flag overrides only applied if flag was explicitly passed (not zero value)
- escape hatch for complex post-load overrides with full flag access

Built-in commands: `version`, `config validate`, `config print`

Global flags: `--config`, `--log-level`, `--log-format`, `--no-color`, `--json`, `--debug`, `--quiet`

Run context propagated via ctx:
- typed config
- output writer
- run flags (debug, quiet, json, no-color)

Error rendering at top level, KindUser vs KindInternal, hints shown.
Exit code mapping, extensible.
Signal handling: SIGTERM/SIGINT → context cancel.

---

## `kleetest`

Test harness for CLI commands. Built on testify.

Run a command with args, captures stdout, stderr, exit code:

```go
result := kleetest.Run(t, app, "config", "validate", "--config", "testdata/good.yaml")

result.ExitCode.Equals(t, 0)
result.Stdout.Contains(t, "valid")
result.Stdout.Empty(t)
result.Stdout.Equals(t, "exact match")
result.Stderr.Contains(t, "error")
result.Stderr.Empty(t)
```

Future: assert on structured JSON output

---

## Future / reserved

- **Doc generation** — markdown/man pages from Go field comments + command tree, AST-based
- **OCS output** — emit an OpenCLI Specification document from command tree + config struct
- **Server helpers** — graceful drain, health/readiness probes
- **`--dry-run` pattern** — helper for commands to declare dry-run support
- **Table action decorator** — declare output shape on command, framework handles table/JSON rendering automatically

---

## Slice 1

Minimal foundation: app boots, config loads, errors surface, version works.

### packages

**errors**
- `Kind`, `Kinder`, `Hinter` — interfaces and constants only, no concrete types

**version**
- `Version`, `Commit`, `BuildDate` vars (ldflags)
- `String()`, `Template`

**xdg**
- XDG path resolution driven by app name

**config**
- struct tags: `yaml`, `env`, `default`, `secret`, `flag`
- precedence: env → project file → user file → system file → defaults
- XDG paths via `xdg` package
- no validation yet

**klee** (root)
- `App[T]`, `New[T]`
- `LoadConfig` — pre-scans args for `--config`, merges layers, `AfterLoad` escape hatch
- `flag:` tag wires config fields to flags, only applied if explicitly set
- global flags: `--config`, `--log-level`, `--log-format`, `--no-color`, `--json`, `--debug`, `--quiet`
- built-in `version` command
- error rendering: `KindUser` → message + hint, `KindInternal` → message + "run with --debug"
- exit code mapping from `Kind`, extensible
- signal handling: SIGTERM/SIGINT → context cancel
- context helpers: `Config[T](ctx)`, `RunFlags(ctx)`

### out of scope for slice 1
- log wiring
- output styling
- `config validate` / `config print` commands
- `kleetest`

### file structure

```
github.com/doron-cohen/klee/
│
├── go.mod
│
├── app.go           — App[T], New[T]
├── context.go       — Config[T], RunFlags, Flags
├── flags.go         — global flag definitions
├── signal.go        — signal handling
│
├── errors/
│   └── errors.go
│
├── version/
│   └── version.go
│
├── xdg/
│   └── xdg.go
│
├── config/
│   └── config.go    — public API
│
└── internal/
    └── config/
        ├── tags.go
        ├── merge.go
        └── loader.go
```
