# Architecture

`gdashlint` is a Go CLI for linting plain Grafana dashboard JSON files with built-in rules and config-defined custom rules.

This document describes the current implementation boundaries and the constraints that should guide near-term changes.

## Product scope

The current implementation supports:

- Plain Grafana dashboard JSON.
- Inputs from files, directories, and stdin for linting.
- File-backed inputs for automatic remediation.
- Built-in rules under the `core.*` namespace.
- YAML configuration with upward discovery.
- Config-defined custom rules.
- Per-path and per-JSONPath ignores.
- Human-readable text output, JSON output, and GitHub Actions annotations.
- CI-friendly exit codes with configurable failure threshold.
- Safe automatic remediations for selected built-in rules.

Out of scope for the current implementation:

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

Cobra also exposes shell completion generation through `gdashlint completion`.

### `gdashlint lint`

Runs enabled rules against one or more dashboards.

```sh
gdashlint lint dashboard.json
gdashlint lint dashboards/
gdashlint lint dashboard.json dashboards/ -
cat dashboard.json | gdashlint lint -
```

File paths are parsed as single dashboard JSON files. Directory paths are walked recursively and include files ending in `.json`. `-` reads one dashboard JSON document from stdin.

### `gdashlint fix`

Applies safe automatic remediations for fixable rules and reports remaining findings after fixes.

```sh
gdashlint fix dashboards/
gdashlint fix dashboards/ --dry-run
gdashlint fix dashboards/ --mode copy
```

`fix` currently requires file-backed dashboard inputs and does not support stdin.

### `gdashlint rules`

Lists available built-in and config-defined rules. Text output is intended for human discovery; JSON output is intended for scripting and documentation generation.

### `gdashlint --version`

Prints build metadata injected by release tooling and exits.

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
  cli.go              # command parsing, flags, help text, process exit mapping

internal/app/
  app.go              # high-level orchestration for lint, fix, and rules commands

internal/config/
  config.go           # config schema, defaults, validation
  discover.go         # upward config discovery
  ignore.go           # suppression filtering

internal/dashboard/
  dashboard.go        # dashboard document representation
  loader.go           # files, directories, stdin
  jsonpath.go         # supported JSONPath-like helper
  write.go            # safe fixed-dashboard writes and copy path helpers

internal/rule/
  finding.go          # finding model
  rule.go             # rule interfaces and fix model
  registry.go         # rule registration and lookup
  severity.go         # severity parsing and ordering

internal/rules/builtin/
  registry.go         # built-in rule set
  ...                 # one file per rule or rule family

internal/rules/config/
  registry.go         # config-defined rule construction
  rules.go            # generic configurable rule implementations

internal/lint/
  runner.go           # applies rules to dashboards and collects findings

internal/output/
  text.go             # human lint output
  json.go             # machine lint output
  github.go           # GitHub Actions annotations
  fix.go              # text fix summaries
  rules.go            # rule metadata output
  sort.go             # severity/file ordering
```

Package responsibilities:

- `cmd/gdashlint`: process entrypoint only.
- `internal/cli`: command syntax, flags, and CLI exit handling; no linting logic.
- `internal/app`: high-level orchestration; converts CLI options into application actions.
- `internal/config`: YAML loading, discovery, defaults, validation, and ignores.
- `internal/dashboard`: input discovery, dashboard parsing, JSONPath-like helpers, and safe writes.
- `internal/rule`: internal domain types for rules, findings, severities, and fixes.
- `internal/rules/builtin`: built-in rules.
- `internal/rules/config`: config-defined custom rules.
- `internal/lint`: rule execution pipeline.
- `internal/output`: deterministic output formatting and ordering.

## Dashboard representation

Grafana dashboards have a large and evolving schema. `gdashlint` avoids modeling the entire dashboard as Go structs.

Dashboards are parsed into a flexible document representation:

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
2. Rules can target arbitrary supported JSON paths without waiting for typed model changes.
3. Config-defined custom rules can be expressed in terms of JSONPath-like paths.

Typed wrappers may be added later for high-value areas, for example dashboard title, tags, variables, panels, and templating. Those wrappers should be convenience helpers over the raw JSON document, not a replacement for it.

## Rule model

Rules are small units that inspect a dashboard and return findings.

```go
type Rule interface {
    Metadata() Metadata
    Check(context.Context, Dashboard) ([]Finding, error)
}

type FixableRule interface {
    Rule
    Fix(context.Context, *Dashboard) ([]Fix, error)
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

A fix contains:

```go
type Fix struct {
    RuleID      string
    File        string
    Path        string
    Description string
}
```

## Severity

Supported severities:

```text
error
warning
info
```

`none` is accepted only as a failure threshold. Severity controls display and CI failure threshold behavior. Configuration may override rule severity.

## Rule ID namespacing

Rule IDs are namespaced.

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
- Makes JSON output more stable for downstream automation.

Current convention:

- Built-in rules use `core.*`.
- Config-defined rules default to `custom.*` when the configured ID does not include a dot.

## Config-defined custom rules

The custom rule mechanism is YAML configuration, not Go code or external processes.

This keeps custom rules:

- Safe to run in CI.
- Easy to review in pull requests.
- Cross-platform.
- Dependency-light.

Supported generic rule types:

- `required`: path must exist.
- `forbidden`: path must not exist.
- `match`: string value must match a regular expression.
- `oneOf`: value must equal one of a configured set.

Config-defined rules may opt into safe declarative fixes with a `fix` block. Supported fix actions are intentionally narrow:

- `set`: set a single target path to a literal YAML value when a `required`, `match`, or `oneOf` rule reports a finding.
- `setDefault`: set a single target path only when it is missing and a `required` rule reports a finding.

`forbidden` rules do not support fixes yet because safe deletion semantics need a separate `remove` action. Custom fix paths default to the rule path, must target one field or array element, and must not contain wildcard selectors. Custom fixes do not compute values, edit multiple paths, call external processes, or infer remediation from dashboard context.

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

Suppressions can match findings by:

- rule ID
- file path glob
- optional JSONPath-like finding path
- optional reason

```yaml
ignore:
  - rule: core.panel-title-required
    paths:
      - dashboards/experimental/**
    reason: Experimental dashboards are noisy while being migrated.
```

Ignores are applied after rule execution and before output. Ignored findings do not count toward exit code failure.

## Output architecture

Output formatting is separated from lint execution.

The lint runner returns a result object independent of output format:

```go
type Result struct {
    Summary  Summary
    Findings []Finding
    Fixes    []Fix
}
```

The output package renders results as text, JSON, or GitHub Actions annotations.

### Default ordering

Default output is sorted globally by severity.

Severity ordering:

```text
error > warning > info
```

Within equal severity, output is sorted by file, then JSON path, then rule ID.

A `--sort file` mode groups findings by file for local debugging.

## JSON output stability

JSON output is a compatibility surface for downstream automation. Breaking changes to field names, rule IDs, or severity values should be treated as compatibility-impacting changes.

Example shape:

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
