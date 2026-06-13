# Configuration

`gdashlint` uses optional YAML configuration for output preferences, failure thresholds, built-in rule overrides, config-defined custom rules, and ignores.

## Discovery

If `--config` is not provided, `gdashlint` searches upward from the current working directory for these files:

1. `.gdashlint.yaml`
2. `.gdashlint.yml`
3. `gdashlint.yaml`
4. `gdashlint.yml`

If none are found, built-in defaults are used.

## Example

```yaml
version: 1

failOn: error

output:
  format: text
  sort: severity

rules:
  core.dashboard-title-required:
    enabled: true
    severity: error

  core.dashboard-uid-required:
    enabled: true
    severity: warning

  core.dashboard-tags-required:
    enabled: true
    severity: warning

  core.refresh-min-interval:
    enabled: true
    severity: error
    options:
      min: 60s

  core.dashboard-not-editable:
    enabled: true
    severity: error

  core.variable-current-empty:
    enabled: true
    severity: warning

  core.panel-title-required:
    enabled: true
    severity: warning

customRules:
  custom.team-owner-tag-format:
    type: match
    severity: error
    path: $.tags[*]
    pattern: "^team:[a-z0-9-]+$"
    message: dashboard must include tags using the team:<name> format

ignore:
  - rule: core.panel-title-required
    paths:
      - dashboards/experimental/**
    reason: Experimental dashboards are migrated gradually.
```

## Top-level fields

### `version`

Configuration schema version.

Current value:

```yaml
version: 1
```

### `failOn`

Minimum severity that causes exit code `1`.

Supported values:

```text
error
warning
info
none
```

Default:

```yaml
failOn: error
```

### `output`

Default output preferences.

```yaml
output:
  format: text
  sort: severity
```

Supported `format` values:

```text
text
json
github
```

Supported `sort` values:

```text
severity
file
```

CLI flags override config values.

### `rules`

Overrides for built-in and registered rules.

```yaml
rules:
  core.dashboard-title-required:
    enabled: true
    severity: error
```

Supported fields:

- `enabled`: enables or disables the rule.
- `severity`: overrides the rule's default severity.
- `options`: rule-specific options supported by selected built-in rules.

Example:

```yaml
rules:
  core.refresh-min-interval:
    enabled: true
    severity: error
    options:
      min: 60s

  core.dashboard-tags-required:
    enabled: true
    severity: warning
    options:
      min: 2
      requiredTags:
        - team:platform
```

Current built-in rule options:

| Rule | Option | Description |
|---|---|---|
| `core.refresh-min-interval` | `min` | Minimum allowed Grafana duration, for example `60s`, `1m`, or `5m`. |
| `core.dashboard-tags-required` | `min` | Minimum number of non-empty dashboard tags. |
| `core.dashboard-tags-required` | `requiredTags` | List of exact tag strings that must be present. |

Unknown rule IDs should be configuration errors unless they are defined under `customRules`.

### `customRules`

Config-defined custom rules.

Custom rule `path` values are validated when configuration is loaded. Invalid JSONPath-like expressions fail fast before any dashboard is linted.

All config-defined custom rules should use stable namespaced IDs. The recommended namespace is `custom.*`. If a custom rule ID does not contain a dot, `gdashlint` prefixes it with `custom.` when registering the rule.

If `severity` is omitted, custom rules default to `warning`. If `message` is omitted, `gdashlint` generates a rule-type-specific message.

Custom rule types are intentionally small and generic.

#### `required`

Requires a JSONPath-like path to exist.

```yaml
customRules:
  custom.dashboard-tags-required:
    type: required
    severity: warning
    path: $.tags
    message: dashboard must define tags
```

#### `forbidden`

Requires a JSONPath-like path not to exist.

```yaml
customRules:
  custom.no-editable-flag:
    type: forbidden
    severity: info
    path: $.editable
    message: dashboard should not set editable explicitly
```

#### `oneOf`

Requires a value to be one of an allowed set. Configure allowed values with `values`.

```yaml
customRules:
  custom.schema-version-known:
    type: oneOf
    severity: warning
    path: $.schemaVersion
    values: [37, 38, 39]
    message: dashboard schemaVersion should be known by this repository
```

#### `match`

Requires a string value to match a regular expression. The `pattern` field is required.

```yaml
customRules:
  custom.dashboard-title-prefix:
    type: match
    severity: warning
    path: $.title
    pattern: "^([A-Z][A-Za-z0-9]+): .+"
    message: dashboard title should start with an owning area prefix
```

### JSONPath-like syntax

Custom rule `path` values and ignore `jsonPaths` use a small JSONPath-like subset.

Supported examples:

```text
$
$.title
$.templating.list[*].current
$.panels[0].title
$.panels[*].gridPos
```

Supported syntax:

- `$` for the root document.
- Dot-separated object fields, for example `$.templating.list`.
- Zero-based array indexes, for example `$.panels[0]`.
- Array wildcards, for example `$.panels[*].title`.

Not supported:

- Recursive descent such as `$..title`.
- Filters such as `$.panels[?(@.type=="row")]`.
- Slices such as `$.tags[0:2]`.
- Quoted field selectors such as `$["title"]`.

### `ignore`

Suppresses findings by rule, file path, and optional JSONPath.

```yaml
ignore:
  - rule: core.panel-title-required
    paths:
      - dashboards/experimental/**
    reason: Experimental dashboards are migrated gradually.
```

Supported fields:

- `rule`: required rule ID.
- `paths`: optional file path glob list.
- `jsonPaths`: optional JSONPath-like finding path list. Values are validated when configuration is loaded.
- `reason`: optional human-readable explanation.

An ignore entry must include at least one `paths` or `jsonPaths` value. When both are present, both dimensions must match: the finding file must match one of the path globs, and the finding JSON path must match one of the configured JSON paths. Ignored findings do not affect output summary failure counts or exit code.

## CLI override precedence

Precedence from highest to lowest:

1. CLI flags.
2. Explicit config file via `--config`.
3. Auto-discovered config file.
4. Built-in defaults.

## Dependency note

YAML support is implemented with:

```text
github.com/goccy/go-yaml
```
