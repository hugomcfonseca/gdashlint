# Examples

## Custom rules

[`custom-rules/gdashlint.yaml`](custom-rules/gdashlint.yaml) demonstrates a practical configuration for plain Grafana dashboard JSON files.

It includes:

- Built-in rule severity overrides.
- Built-in dashboard, refresh, variable, and panel quality rules with rule-specific options.
- A custom allowed-timezone rule.
- A custom dashboard title ownership-prefix rule.
- An example path-based ignore.

Run it against local dashboards:

```sh
go run ./cmd/gdashlint lint dashboards/ --config examples/custom-rules/gdashlint.yaml
```

Run with JSON output:

```sh
go run ./cmd/gdashlint lint dashboards/ \
  --config examples/custom-rules/gdashlint.yaml \
  --format json
```

Run with GitHub Actions annotation output:

```sh
go run ./cmd/gdashlint lint dashboards/ \
  --config examples/custom-rules/gdashlint.yaml \
  --format github
```

Preview safe automatic remediations:

```sh
go run ./cmd/gdashlint lint dashboards/ \
  --config examples/custom-rules/gdashlint.yaml \
  --fix \
  --dry-run
```

Apply fixes to sibling files instead of modifying originals:

```sh
go run ./cmd/gdashlint lint dashboards/ \
  --config examples/custom-rules/gdashlint.yaml \
  --fix \
  --fix-mode copy
```

## Validating against public Grafana dashboards

You can validate the binary against public Grafana dashboard JSON without committing third-party dashboards to this repository:

```sh
mkdir -p /tmp/gdashlint-validation/dashboards
curl -fsSL https://grafana.com/api/dashboards/1860/revisions/37/download \
  -o /tmp/gdashlint-validation/dashboards/node-exporter-full.json
curl -fsSL https://grafana.com/api/dashboards/3662/revisions/2/download \
  -o /tmp/gdashlint-validation/dashboards/prometheus-stats.json

go build -o /tmp/gdashlint-validation/gdashlint ./cmd/gdashlint
/tmp/gdashlint-validation/gdashlint lint \
  /tmp/gdashlint-validation/dashboards \
  --config examples/custom-rules/gdashlint.yaml
```

Use `--fail-on none` when you want to inspect findings without returning a failing exit code:

```sh
/tmp/gdashlint-validation/gdashlint lint \
  /tmp/gdashlint-validation/dashboards \
  --config examples/custom-rules/gdashlint.yaml \
  --format json \
  --fail-on none
```

Preview remediations against the public dashboards:

```sh
/tmp/gdashlint-validation/gdashlint lint \
  /tmp/gdashlint-validation/dashboards \
  --config examples/custom-rules/gdashlint.yaml \
  --fix \
  --dry-run
```

Expected behavior:

- Fixable findings are listed in the fix summary.
- Remaining non-fixable findings are still reported.
- With `--format json`, fix information is included in the `fixes` field.
