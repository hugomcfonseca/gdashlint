# Security Policy

## Supported versions

`gdashlint` is in a pre-1.0 release phase. Security fixes are provided for the latest released version. Older pre-1.0 versions may not receive backported fixes unless maintainers explicitly decide otherwise for a high-impact issue.

## Reporting a vulnerability

Please do not report security vulnerabilities through public GitHub issues.

To report a vulnerability, use GitHub's private vulnerability reporting if enabled for this repository, or contact the maintainers directly.

When reporting, include:

- Affected versions or commits.
- Steps to reproduce the issue.
- Potential impact.
- Any known mitigations.

Maintainers will acknowledge valid reports as soon as practical and coordinate disclosure timelines with reporters.

## Security checks

CI runs Go vulnerability and static security checks, including `govulncheck` and `gosec`. These checks reduce risk but do not replace responsible vulnerability reporting for issues found by manual review or external testing.
