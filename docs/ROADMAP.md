# Roadmap

This roadmap tracks likely next steps after the current pre-1.0 implementation.

## Completed foundation

- Open-source community health files.
- Go CLI with Cobra.
- Dashboard loading from files, directories, and stdin for linting.
- Dedicated `fix` command for safe automatic remediations.
- Diff previews, interactive approval, and `--yes` automation for `gdashlint fix`.
- Built-in rules under the `core.*` namespace.
- Config-defined custom rules with declarative fixes.
- YAML configuration discovery.
- Text, JSON, and GitHub Actions annotation output.
- Minimal Docker image for portable CI usage.
- CI, Dependabot, static analysis, security checks, and tag-based GoReleaser releases.

## Near-term improvements

- Add more focused built-in rules based on real dashboard usage.
- Improve JSONPath support where needed by custom rules.
- Add generated rule documentation from the live registry.
- Add more adoption examples for common dashboard repository layouts.
- Add baseline support for adopting `gdashlint` on repositories with existing findings.

## Remediation improvements

- Consider stdout-based remediation for stdin input.
- Add more safe core remediations where behavior is unambiguous.
- Expand declarative custom-rule remediation actions only when their safety and idempotency constraints are clear.

## Future extensibility

- Evaluate external executable rules if config-defined rules are not expressive enough.
- Revisit Terraform/provider dashboard definitions and Grafana provisioning YAML support.
- Consider public Go APIs only after internal rule and dashboard models stabilize.

## Releases and compatibility

- Continue validating release artifacts and Docker image publishing on early pre-1.0 releases.
- Define compatibility expectations for JSON output, config schema, rule IDs, and fix behavior before a `v1.0.0` release.
