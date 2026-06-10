# Roadmap

This roadmap is intentionally high-level and will evolve as the project design is refined.

## Phase 0: Repository foundation

- Open-source community health files.
- Minimal Go CLI scaffold.
- CI and dependency automation.

## Phase 1: Core design

- Define dashboard input model and supported Grafana dashboard versions.
- Define lint result schema and exit code behavior.
- Define rule interfaces for built-in and custom rules.
- Define configuration file format.
- Capture decisions in [`ARCHITECTURE.md`](ARCHITECTURE.md) and [`CONFIGURATION.md`](CONFIGURATION.md).

## Phase 2: Built-in linting

- Implement dashboard loading and validation.
- Add the initial built-in rule set.
- Add tests and fixtures for representative dashboards.

## Phase 3: Extensibility

- Add custom rule loading strategy.
- Document rule authoring and distribution.
- Add examples for local and CI usage.

## Phase 4: Releases

- Add release packaging.
- Publish installation instructions.
- Define version support and compatibility policy.
