# GitHub Copilot instructions

`gdashlint` is a Go CLI for linting Grafana dashboards with built-in and custom rules.

When suggesting changes:

- Use idiomatic Go and keep public APIs small.
- Do not add dashboard linting implementation details before the project design is established.
- Prefer standard library packages for the CLI scaffold unless a dependency is requested.
- Include tests for new behavior.
- Keep GitHub Actions and repository metadata consistent with open-source best practices.
