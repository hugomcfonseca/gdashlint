# Releasing

Releases are fully automated. Every push to `main` runs `.github/workflows/release.yml`, which calculates the next version from [Conventional Commits](https://www.conventionalcommits.org/), tags the commit, and publishes the release with GoReleaser.

Tooling:

- [semantic-release](https://github.com/semantic-release/semantic-release) calculates the version and pushes the `vMAJOR.MINOR.PATCH` tag. It is configured in `.releaserc.yaml` and uses only the commit analyzer; it does not publish anything itself.
- [GoReleaser](https://goreleaser.com/) builds the binaries, creates the GitHub Release, and pushes the Docker images. It is configured in `.goreleaser.yaml`.

## How a release happens

1. A pull request is merged into `main`.
2. The release workflow runs `go test ./...`.
3. semantic-release inspects the commits since the last `v*` tag:
   - If none of them is releasable, the workflow stops. No tag, release, or image is created.
   - Otherwise it pushes the next `vMAJOR.MINOR.PATCH` tag.
4. When a tag was created, the workflow sets up QEMU and Docker Buildx, logs in to Docker Hub, and runs GoReleaser.
5. GoReleaser publishes the GitHub Release (with a changelog grouped by commit type) and the Docker Hub images.

There is no manual tagging step. Do not push `v*` tags by hand; they are not a release trigger anymore.

## Version calculation

| Commit | Bump |
| --- | --- |
| `fix:`, `perf:` | patch |
| `feat:` | minor |
| `feat!:`, `fix!:`, or a `BREAKING CHANGE:` footer | minor while pre-1.0, major after |
| `docs:`, `test:`, `chore:`, `ci:`, `refactor:`, `style:`, `build:` | no release |

The highest bump among the commits since the last release wins.

Because the project is pre-1.0, breaking changes only bump the minor version. This is set by the `breaking: true` rule in `.releaserc.yaml`, which prevents an accidental jump to `v1.0.0`.

### Releasing v1.0.0

1. Remove the `releaseRules` entry from `.releaserc.yaml` and merge it with a non-releasable commit such as `ci: ...`.
2. Merge a commit with a breaking change (`feat!: ...` or a `BREAKING CHANGE:` footer). The next release will be `v1.0.0`.

### Commit message requirements

Use squash merges with the pull request title as the commit message, or make sure every commit that lands on `main` follows Conventional Commits. Commits that do not follow the convention are ignored when calculating the version.

Dependabot is configured with `fix` for Go modules and `ci` for GitHub Actions, so merged Go dependency updates produce a patch release and GitHub Actions updates do not.

## Artifacts

GoReleaser builds archives for:

- Linux amd64/arm64
- macOS amd64/arm64
- Windows amd64/arm64

Each release includes checksums in `checksums.txt`.

GoReleaser also publishes Docker images to Docker Hub:

- `docker.io/hugomcfonseca/gdashlint:<tag>` such as `v0.1.0`
- `docker.io/hugomcfonseca/gdashlint:latest`

Both tags are multi-platform images for Linux amd64 and arm64.

## Repository requirements

- `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` repository secrets. The token needs permission to push `hugomcfonseca/gdashlint`.
- The workflow pushes tags with the default `GITHUB_TOKEN` (`contents: write`). If a branch protection rule or ruleset on `main` restricts who can push tags, tag creation by `github-actions[bot]` must be allowed.

If the Docker Hub secrets are missing or invalid, the workflow fails after the tag is created and before GoReleaser publishes anything.

## Recovering from a failed release

If the tag was pushed but GoReleaser failed (for example because of Docker Hub credentials), fix the cause and re-run the failed workflow run from the Actions tab. The tag already points at the commit being released, so semantic-release makes no new tag and the workflow resumes at GoReleaser for that tag.

A re-run always uses the original commit, even if later commits have landed on `main`, so the original tag is the one that gets released.
