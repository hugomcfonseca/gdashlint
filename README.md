# gdashlint

[![CI](https://img.shields.io/github/actions/workflow/status/hugomcfonseca/gdashlint/ci.yml?branch=main&label=ci)](https://github.com/hugomcfonseca/gdashlint/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/hugomcfonseca/gdashlint)](https://go.dev/)
[![Latest Release](https://img.shields.io/github/v/release/hugomcfonseca/gdashlint)](https://github.com/hugomcfonseca/gdashlint/releases)
[![License](https://img.shields.io/github/license/hugomcfonseca/gdashlint)](LICENSE)

`gdashlint` is a Go CLI for linting plain Grafana dashboard JSON files with built-in rules, YAML-defined custom rules, CI-friendly output formats, and safe automatic remediations for selected issues.

## Features

- Lint Grafana dashboard JSON from files, directories, or stdin.
- Built-in `core.*` rules for dashboard hygiene, refresh intervals, editability, variables, and panel layout.
- YAML configuration with automatic discovery.
- Config-defined custom rules for common JSONPath-like checks.
- Per-path and per-JSONPath ignores.
- Human-readable, JSON, and GitHub Actions annotation output.
- Rule-specific options for supported built-in rules.
- Safe `fix` command support for selected fixable rules, with diff preview and approval before writes.
- Minimal Docker image for portable CI usage.
- Cross-platform release packaging through GoReleaser.

## Status

The project is in an early pre-1.0 release phase. It supports plain Grafana dashboard JSON only. Terraform/provider dashboard definitions, Grafana provisioning YAML, and external executable/plugin rules are intentionally out of scope for now.

## Install

Install the latest released CLI with Go:

```sh
go install github.com/hugomcfonseca/gdashlint/cmd/gdashlint@latest
```

Or run the Docker image:

```sh
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  lint dashboards/
```

Pre-built archives are available from [GitHub Releases](https://github.com/hugomcfonseca/gdashlint/releases).

To build from source:

```sh
git clone git@github.com:hugomcfonseca/gdashlint.git
cd gdashlint
go build -o bin/gdashlint ./cmd/gdashlint
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

Example text output:

```text
error   manual  core.dashboard-title-required       $.title                  dashboards/api.json              dashboard title is required
warning manual  core.panel-title-required           $.panels[0].title        dashboards/api.json              panel title is required

Found 2 finding(s) in 1 dashboard(s): 1 error(s), 1 warning(s), 0 info(s).
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

  custom.dashboard-timezone:
    type: required
    severity: warning
    path: $.timezone
    message: dashboard timezone should be set
    fix:
      action: setDefault
      value: browser
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

A separate GitHub Action is maintained in `hugomcfonseca/gdashlint-action` for teams that prefer a reusable workflow integration over direct CLI or Docker usage.

## Automatic remediations

Some built-in rules and explicitly configured custom rules are fixable. Preview fixes without writing files:

```sh
gdashlint fix dashboards/ --dry-run
```

For text output, `--dry-run` prints a unified diff preview to stdout, simulates the fixes in memory, reports remaining findings, and exits `1` if any file would be changed.

Apply safe fixes in place:

```sh
gdashlint fix dashboards/
```

Before writing files, text-mode `fix` prints the diff preview and prompts for approval. Use `--yes` for non-interactive automation:

```sh
gdashlint fix dashboards/ --yes
```

Write remediated dashboards to sibling files instead of modifying originals:

```sh
gdashlint fix dashboards/ --mode copy
gdashlint fix dashboards/ --mode copy --suffix .remediated
gdashlint fix dashboards/ --mode copy --yes
```

`fix` currently requires file-backed inputs and does not support stdin. Custom rule fixes are declarative and support rule-type-limited `set` and `setDefault` actions on single-target paths without wildcards.

## Exit codes

```text
0 = command succeeded and no finding met the fail threshold
1 = findings met the fail threshold, fix --dry-run would apply changes, or fix approval was declined
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
go run ./cmd/gdashlint fix dashboards/ --dry-run
```

## Documentation

- [`docs/CLI.md`](docs/CLI.md)
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
