# gdashlint

`gdashlint` is a planned Go CLI for validating Grafana dashboards with built-in lint rules and user-defined extensions.

> Status: initial repository scaffold. The linting engine, rule model, and dashboard parsing commands are intentionally not implemented yet.

## Goals

- Provide sensible built-in lint rules for Grafana dashboard quality and maintainability.
- Support custom rules that teams can define and share.
- Offer a fast CLI suitable for local development and CI pipelines.
- Keep configuration explicit, reviewable, and friendly to open-source collaboration.

## Install

Installation instructions will be added once the first usable release is available.

## Development

Prerequisites:

- Go 1.24 or newer
- `make` (optional, but recommended)

Common commands:

```sh
make fmt
make test
make vet
```

GitHub Actions annotations are available for CI:

```sh
gdashlint lint dashboards/ --format github
```

Safe automatic remediations are available for fixable rules:

```sh
gdashlint lint dashboards/ --fix
gdashlint lint dashboards/ --fix --dry-run
gdashlint lint dashboards/ --fix --fix-mode copy
```

Run the CLI:

```sh
go run ./cmd/gdashlint --version
go run ./cmd/gdashlint rules
go run ./cmd/gdashlint lint dashboards/
```

## Design

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- [`docs/BUILTIN_RULES.md`](docs/BUILTIN_RULES.md)
- [`docs/CONFIGURATION.md`](docs/CONFIGURATION.md)
- [`docs/RELEASING.md`](docs/RELEASING.md)
- [`docs/ROADMAP.md`](docs/ROADMAP.md)

## Contributing

Contributions are welcome. Please read [`CONTRIBUTING.md`](CONTRIBUTING.md), [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md), and [`SECURITY.md`](SECURITY.md) before opening issues or pull requests.

## License

This project is licensed under the MIT License. See [`LICENSE`](LICENSE).
