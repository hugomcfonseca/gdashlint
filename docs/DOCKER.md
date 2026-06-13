# Docker

`gdashlint` publishes a minimal Docker image for running the CLI in CI or other environments where installing a local binary is inconvenient.

## Image

Release images are published to Docker Hub:

```text
docker.io/hugomcfonseca/gdashlint
```

Supported tags:

- `vMAJOR.MINOR.PATCH`
- `latest`

Both tags are multi-platform images for Linux amd64 and arm64. The image uses `scratch`, runs as a non-root numeric user, and contains only the statically linked `gdashlint` binary.

## Usage

Lint dashboards from the current repository:

```sh
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  lint dashboards/
```

Use a config file:

```sh
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  lint dashboards/ --config gdashlint.yaml
```

Emit GitHub Actions annotations from a workflow step:

```sh
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  lint dashboards/ --format github
```

Preview safe automatic remediations without writing files:

```sh
docker run --rm \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  fix dashboards/ --dry-run
```

When applying fixes to mounted files, run the container as your host user so modified files remain writable by you:

```sh
docker run --rm \
  --user "$(id -u):$(id -g)" \
  -v "$PWD:/work" \
  -w /work \
  docker.io/hugomcfonseca/gdashlint:latest \
  fix dashboards/
```

## Building locally

The Dockerfile expects the same platform-specific build context layout used by GoReleaser:

```sh
mkdir -p linux/amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o linux/amd64/gdashlint ./cmd/gdashlint
docker build --build-arg TARGETPLATFORM=linux/amd64 -t gdashlint:local .
rm -rf linux
```

Smoke test the local image:

```sh
docker run --rm gdashlint:local --version
```

## Publishing

Docker Hub publishing is handled by GoReleaser during tag-based releases. The release workflow logs in to Docker Hub with these repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

See [`RELEASING.md`](RELEASING.md) for the full release process.
