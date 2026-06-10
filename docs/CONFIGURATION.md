# Configuration Design

`gdashlint` uses optional YAML configuration.

This document describes the intended first-version configuration model. Exact field names may evolve during implementation, but changes should preserve the same concepts.

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

Initial value:

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

All config-defined custom rules should use stable namespaced IDs. The recommended namespace is `custom.*`.

Initial custom rule types should be intentionally small and generic.

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

Requires a value to be one of an allowed set.

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

Requires a string value to match a regular expression.

```yaml
customRules:
  custom.dashboard-title-prefix:
    type: match
    severity: warning
    path: $.title
    pattern: "^([A-Z][A-Za-z0-9]+): .+"
    message: dashboard title should start with an owning area prefix
```

### `ignore`

Suppresses findings by rule and path.

```yaml
ignore:
  - rule: core.panel-title-required
    paths:
      - dashboards/experimental/**
    reason: Experimental dashboards are migrated gradually.
```

Supported fields:

- `rule`: required rule ID.
- `paths`: file path glob list.
- `jsonPaths`: optional JSONPath-like path list. Values are validated when configuration is loaded.
- `reason`: optional human-readable explanation.

Ignored findings do not affect output summary failure counts or exit code.

## CLI override precedence

Precedence from highest to lowest:

1. CLI flags.
2. Explicit config file via `--config`.
3. Auto-discovered config file.
4. Built-in defaults.

## Dependency note

YAML support requires a non-standard-library dependency. The chosen dependency is:

```text
github.com/goccy/go-yaml
```
