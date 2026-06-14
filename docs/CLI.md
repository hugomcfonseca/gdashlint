# CLI Reference

`gdashlint` is organized around subcommands. Normal command output is written to stdout. Errors, diagnostics, and text-mode fix summaries are written to stderr when needed. Text-mode fix diff previews are written to stdout.

```sh
gdashlint --version
gdashlint lint [paths...] [flags]
gdashlint fix [paths...] [flags]
gdashlint rules [flags]
gdashlint completion [bash|fish|powershell|zsh]
```

## `gdashlint lint`

Runs enabled rules against one or more dashboards.

```sh
gdashlint lint dashboard.json
gdashlint lint dashboards/
gdashlint lint dashboard.json dashboards/ -
cat dashboard.json | gdashlint lint -
```

Flags:

```text
--config path              Path to config file. If omitted, auto-discover config.
--format text|json|github  Output format. Default: text.
--sort severity|file       Finding sort/group mode. Default: severity.
--fail-on error|warning|info|none
                           Minimum severity that returns exit code 1. Default: error.
```

Input rules:

- File paths are parsed as single dashboard JSON files.
- Directory paths are walked recursively and include files ending in `.json`.
- `-` reads one dashboard JSON document from stdin.
- Missing paths, invalid JSON, and unreadable files return exit code `2`.

## `gdashlint fix`

Applies safe automatic remediations for fixable rules and reports remaining findings after fixes.

```sh
gdashlint fix dashboards/ --dry-run
gdashlint fix dashboards/
gdashlint fix dashboards/ --yes
gdashlint fix dashboards/ --mode copy
gdashlint fix dashboards/ --mode copy --suffix .remediated
gdashlint fix dashboards/ --mode copy --yes
```

Flags:

```text
--config path              Path to config file. If omitted, auto-discover config.
--format text|json|github  Output format. Default: text.
--sort severity|file       Finding sort/group mode. Default: severity.
--fail-on error|warning|info|none
                           Minimum severity that returns exit code 1. Default: error.
--mode in-place|copy       Write fixed dashboards in place or to sibling files. Default: in-place.
--suffix .fixed            Suffix used before the extension in copy mode. Default: .fixed.
--dry-run                  Simulate remediations without writing files.
-y, --yes                  Apply fixes without interactive approval.
```

Constraints:

- `fix` requires file-backed dashboard inputs.
- `fix` does not support stdin input yet.
- In copy mode, `--suffix` must be a filename suffix, not a path.
- Fixed files are written with private file permissions.
- Existing symlink destinations are refused to avoid writing through symlinks.
- Without `--dry-run` or `--yes`, `fix` asks for interactive approval before writing. Only `y` or `yes` approves the write.

For text output, `fix --dry-run` writes a "Would apply" fix summary to stderr, writes a unified diff preview to stdout, then writes remaining findings to stdout with a "After simulated fixes" prefix. Without `--dry-run`, text output writes the diff preview to stdout, writes a remaining-findings summary to stdout with an "After fixes" prefix, optionally prompts for approval on stderr, and then writes an "Applied" fix summary to stderr after writes complete. For JSON output, applied or simulated fixes are included under the `fixes` field; diff previews are not emitted.

## `gdashlint rules`

Lists built-in and config-defined rules.

```sh
gdashlint rules
gdashlint rules --format json
gdashlint rules --config gdashlint.yaml
```

Flags:

```text
--config path              Path to config file. If omitted, auto-discover config.
--format text|json|github  Output format. Default: text.
```

`--format json` is useful for scripting or generating rule documentation.

## `gdashlint completion`

Generates shell completion scripts through Cobra.

```sh
gdashlint completion bash
gdashlint completion fish
gdashlint completion powershell
gdashlint completion zsh
```

Use `gdashlint completion <shell> --help` for shell-specific installation instructions.

## Output formats

### `text`

Human-readable output for local use. This is the default format.

### `json`

Machine-readable output for scripts and CI systems.

Example shape:

```json
{
  "summary": {
    "files": 1,
    "findings": 1,
    "errors": 1,
    "warnings": 0,
    "infos": 0
  },
  "findings": [
    {
      "rule_id": "core.dashboard-title-required",
      "severity": "error",
      "message": "dashboard title is required",
      "file": "dashboards/api.json",
      "path": "$.title",
      "fixable": false
    }
  ]
}
```

`fix --format json` may include a `fixes` array.

### `github`

Writes GitHub Actions workflow command annotations. Use it from CI when you want findings to appear inline in pull requests.

```sh
gdashlint lint dashboards/ --format github
```

## Exit codes

```text
0 = command succeeded and no finding met the fail threshold
1 = lint/fix completed and at least one finding met the fail threshold, fix --dry-run would apply changes, or fix approval was declined
2 = usage, configuration, input, parse, or runtime error
```
