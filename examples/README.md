# Examples

## Custom rules

[`custom-rules/gdashlint.yaml`](custom-rules/gdashlint.yaml) demonstrates user-defined rules that can be applied to plain Grafana dashboard JSON files.

Run it against local dashboards:

```sh
go run ./cmd/gdashlint lint dashboards/ --config examples/custom-rules/gdashlint.yaml
```

Run with JSON output:

```sh
go run ./cmd/gdashlint lint dashboards/ --config examples/custom-rules/gdashlint.yaml --format json
```

Run with GitHub Actions annotation output:

```sh
go run ./cmd/gdashlint lint dashboards/ --config examples/custom-rules/gdashlint.yaml --format github
```

The example config includes:

- Built-in rule severity overrides.
- Built-in dashboard, refresh, variable, and panel quality rules with rule-specific options.
- A custom allowed-timezone rule.
- A custom title naming convention.
- An example path-based ignore.

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
