# Development Guide

## Requirements

- Go 1.27 or newer (`go version`). GNU Make and Bash for the Make targets; each target below also
  shows the plain `go` command it runs.
- A C compiler only for `make test-race`; builds and every other gate use `CGO_ENABLED=0`.
- Network access the first time, to download the pinned tools into the module cache.

## First run

```bash
make ci                  # every required gate
make build               # bin/agent-go
bin/agent-go version
make plan-status         # the next eligible task
```

## Make targets

| Target | Runs | Purpose |
| --- | --- | --- |
| `make build` | `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/ ./cmd/agent-go` | Static host binary |
| `make build-all` | the same for each `PLATFORMS` entry | `bin/agent-go-<os>-<arch>[.exe]` |
| `make run` | `go run ./cmd/agent-go $(ARGS)` | Run from source; `ARGS` defaults to `version` |
| `make fmt` | `gofmt -s -w` | Format |
| `make fmt-check` | `gofmt -s -l` | Fail on unformatted files |
| `make vet` | `go vet ./...` | Compiler-adjacent checks |
| `make lint` | `go tool staticcheck ./...` | Static analysis |
| `make tidy-check` | `go mod tidy -diff` | Fail when `go.mod` or `go.sum` is stale |
| `make test` | `go test ./...` | Unit, integration and tool tests |
| `make test-race` | `CGO_ENABLED=1 go test -race ./...` | Race detector; CI runs it after `make ci` |
| `make coverage` | `go test -coverprofile=coverage.out ./...` | Diagnostic report, never a gate |
| `make ci` | fmt-check, vet, lint, tidy-check, test, build-all, lint-spec-plan, lint-doc-links | Required before accepting a task |
| `make dist` | `build-all`, then `go run ./tools/dist ... $(PLATFORMS)` | Archives and `SHA256SUMS` in `dist/`; `DIST_VERSION=vX.Y.Z` requires that version |
| `make plan-status` | `go run ./tools/plancheck status` | Open task counts and the next eligible task |
| `make lint-spec-plan` | `go run ./tools/plancheck lint` | Registry, plan and records agree |
| `make lint-doc-links` | `go run ./tools/plancheck links` | Relative Markdown links and anchors resolve |
| `make clean` | `rm -rf bin dist coverage.out` | Remove outputs |

## Tools and dependencies

Development tools are pinned with the `tool` directive in `go.mod` and run with `go tool <name>`,
so every machine and CI use the same version without a global install. Add one with
`go get -tool <package>@<version>` and list it in the
[dependency table](../openspec/spec.md#dependencies). Upgrade with `go get -tool <package>@latest`
followed by `go mod tidy`, and commit `go.mod` and `go.sum` together.

## Versioning and releases

Versions are Git tags `vMAJOR.MINOR.PATCH` ([Semantic Versioning](https://semver.org); `v0.y.z`
before 1.0). `go build` stamps the tag, or a pseudo-version for an untagged commit, into the
binary, and `agent-go version` reports it; nothing in the tree records the version.

To release, tag a commit that passed CI and push the tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow rebuilds from a clean checkout of the tag and publishes the archives and
`SHA256SUMS` as a GitHub release. `make dist` gives the same archives locally; its names carry a
pseudo-version (with `+dirty` for uncommitted changes) until the commit is tagged.

## CI

`.github/workflows/ci.yml` runs `make ci`, `make test-race` and `make dist` on `ubuntu-latest`
for pushes to `main` and for pull requests, with the Go version from `go.mod`.
`.github/workflows/release.yml` runs on `v*.*.*` tags: `make ci`, `make dist` with
`DIST_VERSION` set to the tag, a checksum check, and `gh release create`. Keep workflows thin
wrappers over Make targets so local and CI results agree.
