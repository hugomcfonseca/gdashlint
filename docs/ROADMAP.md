# Roadmap

This roadmap tracks likely next steps after the first functional implementation.

## Completed foundation

- Open-source community health files.
- Go CLI with Cobra.
- Dashboard loading from files, directories, and stdin.
- Built-in rules under the `core.*` namespace.
- Config-defined custom rules.
- YAML configuration discovery.
- Text, JSON, and GitHub Actions annotation output.
- Safe automatic remediations for selected rules.
- Minimal Docker image for portable CI usage.
- CI, Dependabot, and tag-based GoReleaser releases.

## Near-term improvements

- Add more focused built-in rules based on real dashboard usage.
- Improve JSONPath support where needed by custom rules.
- Add generated rule documentation from the live registry.
- Add CI examples for using `--format github` in downstream dashboard repositories.
- Add baseline support for adopting `gdashlint` on repositories with existing findings.

## Remediation improvements

- Add a diff preview mode for `--fix`.
- Consider stdout-based remediation for stdin input.
- Add more safe core remediations where behavior is unambiguous.
- Explore declarative custom-rule remediations after the core remediation model stabilizes.

## Future extensibility

- Evaluate external executable rules if config-defined rules are not expressive enough.
- Revisit Terraform/provider dashboard definitions and Grafana provisioning YAML support.
- Consider public Go APIs only after internal rule and dashboard models stabilize.

## Releases and compatibility

- Publish the first release once the CLI surface is considered stable enough for early users.
- Define compatibility expectations for JSON output, config schema, and rule IDs before a `v1.0.0` release.
