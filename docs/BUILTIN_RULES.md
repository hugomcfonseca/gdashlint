# Built-in Rules

Built-in rule IDs use the `core.*` namespace.

| Rule | Default severity | Fixable | Description |
|---|---:|:---:|---|
| `core.dashboard-title-required` | `error` | No | Dashboard `title` must be non-empty. |
| `core.dashboard-uid-required` | `warning` | No | Dashboard `uid` should be non-empty and stable. |
| `core.dashboard-tags-required` | `warning` | No | Dashboard `tags` should contain at least one non-empty tag. Supports `min` and `requiredTags` options. |
| `core.refresh-min-interval` | `error` | Yes | Dashboard `refresh` must be disabled/empty or at least `60s`. Supports the `min` option. |
| `core.dashboard-not-editable` | `error` | Yes | Dashboard `editable`, when present, must be `false`. |
| `core.variable-current-empty` | `warning` | Yes | Dashboard variables must not persist non-empty `current` selections. |
| `core.panel-title-required` | `warning` | No | Panels should define non-empty `title` values. |
| `core.panel-type-required` | `warning` | No | Panels should define non-empty `type` values. |
| `core.panel-grid-position-required` | `warning` | No | Panels should define `gridPos`. |
| `core.panel-grid-position-valid` | `error` | No | Panel `gridPos` values must be valid for Grafana's 24-column grid. |

## Automatic remediations

Safe automatic remediations are available through `gdashlint lint --fix`.

Supported fixable rules:

- `core.dashboard-not-editable`: sets `editable` to `false` when explicitly `true`.
- `core.refresh-min-interval`: raises `refresh` to the configured minimum.
- `core.variable-current-empty`: clears persisted variable `current` selections.

Examples:

```sh
gdashlint lint dashboards/ --fix
gdashlint lint dashboards/ --fix --dry-run
gdashlint lint dashboards/ --fix --fix-mode copy
gdashlint lint dashboards/ --fix --fix-mode copy --fix-suffix .remediated
```

`--fix` currently requires file-backed inputs and does not support stdin.

When `--fix` is used, lint output is based on the post-fix dashboard state. With `--dry-run`, fixes are simulated in memory and files are not written. `--fix --dry-run` exits with code `1` when any fix would be applied, even if no findings would remain after the simulated fixes.

For text output, fix summaries are written as a human-readable section before remaining findings. For structured output such as `--format json`, fixes are included in the output payload under `fixes` instead of being written as separate text.

## Notes

### Refresh intervals

`core.refresh-min-interval` accepts Grafana-style durations such as:

- `60s`
- `1m`
- `5m`
- `1h`

Intervals lower than 60 seconds fail. Empty or missing `refresh` is treated as automatic refresh disabled.

The threshold can be configured:

```yaml
rules:
  core.refresh-min-interval:
    options:
      min: 5m
```

### Dashboard tags

`core.dashboard-tags-required` can enforce a minimum number of tags and exact required tags:

```yaml
rules:
  core.dashboard-tags-required:
    options:
      min: 2
      requiredTags:
        - team:platform
```

### Editable dashboards

`core.dashboard-not-editable` allows missing `editable` values and explicit `false` values. Explicit `true` is reported.

### Variable current values

`core.variable-current-empty` checks `$.templating.list[*].current` and reports persisted selections such as:

```json
{
  "current": {
    "text": "prod",
    "value": "prod"
  }
}
```

Empty objects and empty strings/arrays are allowed.

### Panel traversal

Panel rules inspect top-level panels and nested row panels under `panels[*].panels[*]`.

### Grid position validation

`core.panel-grid-position-valid` requires:

- `w > 0`
- `h > 0`
- `x >= 0`
- `y >= 0`
- `x + w <= 24`
