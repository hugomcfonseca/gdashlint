# gdashlint

`gdashlint` is a Go CLI for linting plain Grafana dashboard JSON files with built-in rules, YAML-defined custom rules, CI-friendly output formats, and safe automatic remediations for selected issues.

## Features

- Lint Grafana dashboard JSON from files, directories, or stdin.
- Built-in `core.*` rules for dashboard hygiene, refresh intervals, editability, variables, and panel layout.
- YAML configuration with automatic discovery.
- Config-defined custom rules for common JSONPath-based checks.
- Per-path and per-JSONPath ignores.
- Human-readable, JSON, and GitHub Actions annotation output.
- Rule-specific options for supported built-in rules.
- Safe `--fix` support for selected fixable rules.
- Minimal Docker image for portable CI usage.
- Cross-platform release packaging through GoReleaser.

## Status

The project has a functional first implementation. It supports plain Grafana dashboard JSON only. Terraform/provider dashboard definitions, Grafana provisioning YAML, and external executable/plugin rules are intentionally out of scope for now.

## Install

Until the first GitHub release is published, build from source:

```sh
git clone git@github.com:hugomcfonseca/gdashlint.git
cd gdashlint
go build -o bin/gdashlint ./cmd/gdashlint
```

After releases are available, download archives from GitHub Releases, install with Go, or run the Docker image:

```sh
go install github.com/hugomcfonseca/gdashlint/cmd/gdashlint@latest

docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  lint dashboards/
```

## Quick start

Run against one dashboard:

```sh
gdashlint lint dashboard.json
```

Run recursively against a directory:

```sh
gdashlint lint dashboards/
```

Read from stdin:

```sh
cat dashboard.json | gdashlint lint -
```

List available rules:

```sh
gdashlint rules
gdashlint rules --format json
```

Print version information:

```sh
gdashlint --version
```

## Configuration

`gdashlint` works without config, using built-in defaults. If `--config` is not provided, it searches upward from the current directory for:

1. `.gdashlint.yaml`
2. `.gdashlint.yml`
3. `gdashlint.yaml`
4. `gdashlint.yml`

Minimal example:

```yaml
version: 1

failOn: error

output:
  format: text
  sort: severity

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
      min: 1

customRules:
  custom.dashboard-title-prefix:
    type: match
    severity: info
    path: $.title
    pattern: "^[A-Za-z][A-Za-z0-9 _-]+: .+"
    message: dashboard title should start with an ownership prefix
```

See [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md) for the full schema.

## Output formats

Human-readable text is the default:

```sh
gdashlint lint dashboards/
```

JSON output is available for scripts and CI systems:

```sh
gdashlint lint dashboards/ --format json
```

GitHub Actions annotations are available for PR checks:

```sh
gdashlint lint dashboards/ --format github
```

## Automatic remediations

Some built-in rules are fixable. Apply safe fixes in place:

```sh
gdashlint lint dashboards/ --fix
```

Preview fixes without writing files:

```sh
gdashlint lint dashboards/ --fix --dry-run
```

Write remediated dashboards to sibling files instead of modifying originals:

```sh
gdashlint lint dashboards/ --fix --fix-mode copy
gdashlint lint dashboards/ --fix --fix-mode copy --fix-suffix .remediated
```

`--fix` currently requires file-backed inputs and does not support stdin. With `--dry-run`, fixes are simulated in memory and the command exits `1` if any file would be changed.

## Exit codes

```text
0 = command succeeded and no finding met the fail threshold
1 = findings met the fail threshold, or --fix --dry-run would apply changes
2 = usage, configuration, input, parse, or runtime error
```

The default fail threshold is `error`. Override it with:

```sh
gdashlint lint dashboards/ --fail-on warning
gdashlint lint dashboards/ --fail-on none
```

## Development

Prerequisites:

- Go 1.26 or newer
- `make` optional, but recommended

Common commands:

```sh
make fmt
make test
make vet
make check
```

Run from source:

```sh
go run ./cmd/gdashlint --version
go run ./cmd/gdashlint rules
go run ./cmd/gdashlint lint dashboards/
```

## Documentation

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- [`docs/BUILTIN_RULES.md`](docs/BUILTIN_RULES.md)
- [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md)
- [`docs/DOCKER.md`](docs/DOCKER.md)
- [`docs/RELEASING.md`](docs/RELEASING.md)
- [`docs/ROADMAP.md`](docs/ROADMAP.md)
- [`examples/README.md`](examples/README.md)

## Contributing

Contributions are welcome. Please read [`CONTRIBUTING.md`](CONTRIBUTING.md), [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md), and [`SECURITY.md`](SECURITY.md) before opening issues or pull requests.

## License

This project is licensed under the MIT License. See [`LICENSE`](LICENSE).
