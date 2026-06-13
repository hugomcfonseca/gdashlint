# Architecture

`gdashlint` is a Go CLI for linting plain Grafana dashboard JSON files with built-in rules and config-defined custom rules.

This document captures the current high-level architecture and the design boundaries for near-term development.

## Product scope

The current implementation supports:

- Plain Grafana dashboard JSON.
- Inputs from files, directories, and stdin.
- Built-in rules.
- YAML configuration.
- Config-defined custom rules.
- Per-path ignores and suppressions.
- Human-readable and machine-readable output.
- CI-friendly exit codes with configurable failure threshold.

Out of scope for the first functional version:

- Terraform/provider dashboard definitions.
- Grafana provisioning YAML.
- Grafana dashboard export bundles beyond plain dashboard JSON.
- Go plugin rules.
- External executable rules.
- A public Go API.

## CLI shape

The CLI uses subcommands for primary behavior and keeps version as a flag.

```sh
gdashlint --version
gdashlint lint [paths...] [flags]
gdashlint fix [paths...] [flags]
gdashlint rules [flags]
```

### `gdashlint lint`

Runs enabled rules against one or more dashboards.

Initial flags:

```sh
--config path              Path to config file. If omitted, auto-discover config.
--format text|json|github  Output format. Default: text.
--sort severity|file       Finding sort/group mode. Default: severity.
--fail-on error|warning|info|none
                           Minimum severity that returns exit code 1. Default: error.
```

Inputs:

```sh
gdashlint lint dashboard.json
gdashlint lint dashboards/
gdashlint lint dashboard.json dashboards/ -
cat dashboard.json | gdashlint lint -
```

Input rules:

- File paths are parsed as single dashboard JSON files.
- Directory paths are walked recursively.
- Directory traversal initially includes files ending in `.json`.
- `-` reads one dashboard JSON document from stdin.
- Missing paths, invalid JSON, and unreadable files are runtime errors.

### `gdashlint fix`

Applies safe automatic remediations for fixable rules and reports remaining findings after fixes.

```sh
--config path              Path to config file. If omitted, auto-discover config.
--format text|json|github  Output format. Default: text.
--sort severity|file       Finding sort/group mode. Default: severity.
--fail-on error|warning|info|none
                           Minimum severity that returns exit code 1. Default: error.
--mode in-place|copy       Write fixed dashboards in place or to sibling files.
--suffix .fixed            Suffix used before .json in copy mode.
--dry-run                  Simulate remediations without writing files.
```

`fix` currently requires file-backed dashboard inputs and does not support stdin.

### `gdashlint rules`

Lists available built-in and config-defined rules.

The command should be useful for discovery and CI debugging. JSON output can be added through the shared `--format` behavior if useful.

### `gdashlint --version`

Prints build metadata and exits.

## Exit codes

```text
0 = command succeeded and no finding met the fail threshold
1 = lint/fix completed and at least one finding met the fail threshold, or fix --dry-run would apply changes
2 = usage, configuration, input, parse, or runtime error
```

Warnings do not fail CI by default. The default failure threshold is `error`.

## Internal package layout

All implementation packages remain under `internal` until the extension model is proven.

```text
cmd/gdashlint/
  main.go

internal/cli/
  cli.go              # command parsing, flags, help text

internal/app/
  app.go              # wires config, loaders, registry, runner, output

internal/config/
  config.go           # config schema, defaults, validation
  discover.go         # upward config discovery

internal/dashboard/
  dashboard.go        # dashboard document representation
  loader.go           # files, directories, stdin
  jsonpath.go         # JSON path helpers, if needed

internal/rule/
  finding.go          # finding, location, severity
  rule.go             # rule interfaces
  registry.go         # rule registration and lookup

internal/rules/builtin/
  registry.go         # built-in rule set
  ...                 # one file per rule or rule family

internal/rules/config/
  registry.go         # config-defined rule construction
  ...                 # generic configurable rule implementations

internal/lint/
  runner.go           # applies rules to dashboards and collects findings

internal/output/
  text.go             # human output
  json.go             # machine output
  sort.go             # severity/file ordering
```

Package responsibilities:

- `cmd/gdashlint`: process entrypoint only.
- `internal/cli`: command syntax and flag parsing; no linting logic.
- `internal/app`: high-level orchestration; converts CLI options into actions.
- `internal/config`: YAML loading, discovery, defaults, and validation.
- `internal/dashboard`: input discovery and dashboard parsing.
- `internal/rule`: stable internal domain types for rules and findings.
- `internal/rules/builtin`: built-in rules.
- `internal/rules/config`: config-defined custom rules.
- `internal/lint`: rule execution pipeline.
- `internal/output`: output formatting and ordering.

## Dashboard representation

Grafana dashboards have a large and evolving schema. The first version should avoid modeling the entire dashboard as Go structs.

Instead, parse dashboards into a flexible document representation:

```go
type Dashboard struct {
    Source Source
    Root   any
}

type Source struct {
    Name  string
    Path  string
    Stdin bool
}
```

Rules inspect `Root`, normally a `map[string]any`, with helper functions for common JSON access patterns.

This provides three advantages:

1. The tool can handle Grafana fields it does not know about yet.
2. Rules can target arbitrary JSON paths without waiting for typed model changes.
3. Config-defined custom rules can be expressed in terms of JSON paths.

Typed wrappers can be added later for high-value areas, for example dashboard title, tags, variables, panels, and templating. Those wrappers should be convenience helpers over the raw JSON document, not a replacement for it.

## Rule model

Rules are small units that inspect a dashboard and return findings.

Conceptual interface:

```go
type Rule interface {
    Metadata() Metadata
    Check(context.Context, Dashboard) ([]Finding, error)
}

type Metadata struct {
    ID          string
    Name        string
    Description string
    Severity    Severity
    Source      SourceKind
    Fixable     bool
}
```

A finding contains:

```go
type Finding struct {
    RuleID   string
    Severity Severity
    Message  string
    File     string
    Path     string // JSONPath-like path, e.g. $.panels[0].title
    Fixable  bool
}
```

### Severity

Supported severities:

```text
error
warning
info
```

Severity controls display and CI failure threshold behavior. Configuration may override rule severity.

## Rule ID namespacing

Rule IDs should be namespaced.

Examples:

```text
core.dashboard-title-required
core.panel-title-required
custom.team-owner-required
custom.no-production-refresh-under-30s
```

Advantages:

- Avoids collisions between built-in and user-defined rules.
- Makes ownership clear in output and config.
- Lets future rule packs use their own namespace, e.g. `grafana.*`, `prometheus.*`, `company.*`.
- Allows broad config targeting later, e.g. enabling/disabling `custom.*`.
- Makes JSON output more stable for downstream automation.

Initial convention:

- Built-in rules use `core.*`.
- Config-defined rules default to `custom.*`, unless explicitly prefixed by the user.

## Config-defined custom rules

The first custom rule mechanism is YAML configuration, not Go code or external processes.

This keeps the first implementation:

- Safe to run in CI.
- Easy to review in pull requests.
- Cross-platform.
- Dependency-light.
- Compatible with future external rule support.

Initial generic rule types should focus on common JSON checks, such as:

- required path exists
- path must not exist
- string matches regex
- array minimum/maximum length
- value equals one of a set

The exact rule set can be implemented incrementally after the core runner exists.

## Config discovery

Config is optional. If no config is found, defaults apply.

Discovery behavior:

1. If `--config` is provided, load that exact file.
2. Otherwise, search from the current working directory upward for:
   - `.gdashlint.yaml`
   - `.gdashlint.yml`
   - `gdashlint.yaml`
   - `gdashlint.yml`
3. Stop at the filesystem root.
4. If no file is found, use default configuration.

Invalid explicit config is an error. Missing auto-discovered config is not an error.

## Ignores and suppressions

The first version should support per-path ignores in config.

Suppression dimensions:

- rule ID
- path glob
- optional JSONPath
- optional reason

Conceptual example:

```yaml
ignore:
  - rule: core.panel-title-required
    paths:
      - dashboards/experimental/**
    reason: Experimental dashboards are noisy while being migrated.
```

Recommended behavior:

- Ignores should be applied after rule execution and before output.
- Ignored findings should not count toward exit code failure.
- A later version may add `--show-ignored` for auditability.

## Output architecture

Output formatting is separated from lint execution.

The lint runner returns a result object independent of output format:

```go
type Result struct {
    Findings []Finding
    Summary  Summary
}
```

The output package renders the result as `text`, `json`, or GitHub Actions annotations.

### Default ordering

Default output is sorted globally by severity.

Severity ordering:

```text
error > warning > info
```

Within equal severity, sort by file, then JSON path, then rule ID.

A `--sort file` mode groups findings by file for local debugging.

## JSON output stability

JSON output should be treated as a compatibility surface once released.

Initial shape:

```json
{
  "summary": {
    "files": 1,
    "findings": 1,
    "errors": 1,
    "warnings": 0,
    "infos": 0
  },
  "findings": [
    {
      "rule_id": "core.dashboard-title-required",
      "severity": "error",
      "message": "dashboard title is required",
      "file": "dashboards/api.json",
      "path": "$.title",
      "fixable": false
    }
  ],
  "fixes": [
    {
      "rule_id": "core.dashboard-not-editable",
      "file": "dashboards/api.json",
      "path": "$.editable",
      "description": "set editable to false"
    }
  ]
}
```

## Implementation sequence

Recommended first implementation slices:

1. CLI command shape and option parsing.
2. Core domain types: severity, finding, dashboard, rule metadata.
3. Config loading and discovery with defaults.
4. Dashboard input loader for files, directories, and stdin.
5. Lint runner and registry with no-op/empty rule set support.
6. Output rendering for text and JSON.
7. Built-in rule registry.
8. First built-in rules.
9. Config-defined rule types.
10. Ignores/suppressions.

This order keeps the architecture testable before rule behavior grows.
