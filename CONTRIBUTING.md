# Contributing

Thanks for your interest in contributing to `gdashlint`.

## Prerequisites

- Go version compatible with `go.mod`.
- `make`, optional but recommended.
- Docker, only needed when changing Docker or release-image behavior.

## Development workflow

1. Fork the repository and create a feature branch.
2. Keep changes focused and include tests for behavior changes.
3. Update documentation when behavior, configuration, CLI output, or workflows change.
4. Run the narrowest relevant local checks before opening a pull request.
5. Open a pull request using the provided template.

Common local commands:

```sh
make fmt
make test
make vet
make check
```

Useful CI-equivalent checks:

```sh
go test -race ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run
go run honnef.co/go/tools/cmd/staticcheck@2026.1 ./...
go run github.com/securego/gosec/v2/cmd/gosec@v2.27.1 ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

If a check is not relevant to your change or cannot run locally, note that in the pull request.

## Commit and pull request expectations

- Use clear, descriptive commit messages.
- Explain the motivation and user impact of the change.
- Keep unrelated refactors out of feature or bug-fix pull requests.
- Update tests for behavior changes.
- Update documentation for user-facing behavior, configuration, or workflow changes.

## Reporting issues

Please use the GitHub issue templates when reporting bugs or proposing features. Include enough detail for maintainers to reproduce or evaluate the request.

For security issues, follow [`SECURITY.md`](SECURITY.md) instead of opening a public issue.
