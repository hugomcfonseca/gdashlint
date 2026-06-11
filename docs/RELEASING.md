# Releasing

Releases are built and published by GoReleaser through `.github/workflows/release.yml` and `.goreleaser.yaml`.

## Current release model

Releases are tag-only. Pushing a semver-like tag triggers the release workflow.

Supported release tags use this format:

```text
vMAJOR.MINOR.PATCH
```

Examples:

```text
v0.1.0
v1.2.3
v1.2.3-rc.1
```

## Creating a release

Create and push an annotated tag:

```sh
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

The tag push triggers the GoReleaser workflow, which will:

1. Check out the release tag.
2. Run tests.
3. Set up QEMU and Docker Buildx for multi-platform image publishing.
4. Log in to Docker Hub.
5. Run GoReleaser.
6. Publish release artifacts to GitHub Releases and Docker Hub.

## Version calculation options

There are two common approaches.

### Manual versioning, current setup

Maintainers choose the version number when creating the release tag.

Pros:

- Simple.
- Predictable.
- The Git tag is the single source of truth.
- No commit-message convention required.
- Good while the project is still early.

Cons:

- Maintainers must decide each version.
- No automatic changelog/version calculation beyond GoReleaser's release changelog.

### Automatic versioning, possible future setup

A tool such as Release Please can calculate the next version from Conventional Commits and open a release PR.

Example mapping:

- `fix:` increments patch.
- `feat:` increments minor.
- `feat!:` or `BREAKING CHANGE:` increments major.

Pros:

- Consistent version calculation.
- Maintainer reviews the release PR before the tag is created.
- Works well once commit conventions are established.

Cons:

- Requires Conventional Commits discipline.
- Adds another workflow/tool.
- Can be noisy for young projects.

Recommendation: keep manual tag-based versioning for the initial releases. Add Release Please later once the project has stable contribution conventions.

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

## Docker Hub credentials

The release workflow requires these repository secrets:

- `DOCKERHUB_USERNAME`: Docker Hub username.
- `DOCKERHUB_TOKEN`: Docker Hub access token with permission to push `hugomcfonseca/gdashlint`.

If either secret is missing or invalid, the release workflow will fail before GoReleaser publishes artifacts.
