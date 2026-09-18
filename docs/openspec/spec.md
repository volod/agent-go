# agent-go Specification

## Purpose

`agent-go` is a copy-ready template for Go command-line projects that people and coding agents
develop together. A fresh copy builds, tests and plans on day one: it ships a static `agent-go`
command, one set of agent rules, quality gates shared by local work and CI, and a planning
workflow in which every change traces from a specified capability to tests and current-state
documentation.

This specification is living. Behavior, boundaries and evaluation belong here, in capability
sections below or in capability pages copied from the [page template](template.md). Remaining work
belongs in the [plan](../impl/plan.md); delivered behavior in [current state](../impl/current.md);
package boundaries in the [architecture](architecture.md). A need discovered during
implementation enters here first, through [capability changes](../guide/planning-workflow.md#capability-changes).

## Scope

In scope:

- a Go module in the standard layout (`cmd/`, `internal/`, `tools/`, `test/`) that builds one
  static executable for Linux, macOS and Windows;
- reproducible release archives with checksums, published to GitHub Releases from a version tag;
- Make entry points that wrap plain `go` commands and are shared by local work and CI;
- gates for formatting, `go vet`, staticcheck, module tidiness, tests, the race detector and
  planning integrity;
- `AGENTS.md` as the only rule source, with adapters for Claude, Gemini and Cursor;
- a capability registry, forward plan, task records and current-state pages.

Out of scope until a specified capability needs them: a business domain, frameworks, runtime
dependencies, configuration files, persistence, network services, containers, package-manager
publishing, code signing and deployment.

## Design principles

1. **Standard library first.** Every dependency is justified in the [dependency table](#dependencies).
2. **Thin edges.** `main` wires the process, `internal/cli` owns the command-line contract, and
   domain packages return errors to it.
3. **One workflow.** Make targets wrap plain `go` commands; CI runs the same targets.
4. **Deterministic evidence.** Tests are network-free and repeatable; acceptance is proven by
   commands another contributor can rerun.
5. **Traceable work.** Specification, plan, record and current-state page agree, and
   `make lint-spec-plan` checks it.

## Cross-cutting rules

### Platforms and build

- Go 1.27 or newer; module `github.com/volod/agent-go`; binary `agent-go` (`agent-go.exe`).
- Builds use `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w"`. `make build` targets the host;
  `make build-all` cross-builds `linux/amd64`, `linux/arm64`, `darwin/arm64` and `windows/amd64`.
- The version is the Go toolchain's VCS stamp read through `runtime/debug`: a tagged commit
  reports its tag (`v1.2.3`), other builds a pseudo-version with `+dirty` for local changes, and
  `go run` or `go test` report `devel`. Releases are Git tags `vMAJOR.MINOR.PATCH`; see
  [release distribution](#release-distribution).
- Gates run on Linux. Other platforms are cross-compiled; platform-specific code lives in
  build-tagged files.

### Dependencies

The product has no runtime dependencies. Development tools are pinned with the `tool` directive
in `go.mod`, run with `go tool`, and are never linked into the binary.

| Dependency | Kind | Use |
| --- | --- | --- |
| [`honnef.co/go/tools/cmd/staticcheck`](https://staticcheck.dev) | tool | Static analysis in `make lint` |

Adding a row is a specification change: name the need, the standard-library alternative that was
rejected and why.

### Command-line contract

- Synopsis: `agent-go <command> [arguments]`. Commands: `help`, `version`.
- stdout carries command results only; usage, errors and logs go to stderr. Logs use `log/slog`.
- Exit codes: `0` success, `1` failure, `2` usage error, `130` interrupted by SIGINT or SIGTERM.

## Project foundation

The repository builds, tests and checks itself from a fresh clone with only Go and Make.
`make ci` runs every required gate, including cross-builds for every platform, and GitHub Actions
runs the same target plus the race detector and `make dist`.

Boundary: the foundation owns layout, tooling and gates, not product behavior.

Evaluation: `make ci` passes on a fresh clone; the planning tool's tests reproduce each defect it
detects; the black-box test drives the built binary through its exit codes. A negative result
names the failing gate and never weakens it.

## Release distribution

`make dist` packages the binary of every build platform into
`agent-go-<version>-<os>-<arch>.tar.gz` (`.zip` for Windows). Each archive holds one directory of
the same name containing the executable, `README.md` and `LICENSE`. `dist/SHA256SUMS` lists every
archive in `sha256sum -c` format. The version in the names is the one stamped into the binaries,
so an archive never disagrees with `agent-go version`. Archives are reproducible: the same commit
and Go toolchain give the same bytes, because entry order, modes and owners are fixed and every
entry time is the commit time.

Pushing a tag `vMAJOR.MINOR.PATCH` publishes a GitHub release: the release workflow runs
`make ci`, builds the archives with `DIST_VERSION` set to the tag, verifies the checksums and
uploads the archives and `SHA256SUMS` with generated notes. `make dist` fails before writing any
archive when a binary reports another version, binaries come from different builds, or a binary
has no module version.

Boundary: no signing, SBOM, container images, installers or package-manager publishing, and no
version bump in the tree; the tag is the version.

Evaluation: packaging tests check archive names, members, modes, times and checksums, byte-for-byte
reproducibility and every refusal; CI runs `make dist` on every push and pull request. A negative
result keeps the previous release and names the refused binary.

## Project identity

`agent-go version` prints one line naming the command, version, module path, Go version and
platform, for example `agent-go v0.1.0 (github.com/volod/agent-go, go1.27.1, linux/amd64)`. After
a repository is created from the template, its module path, command name, README and this
specification name and describe the new product.

Boundary: identity only. It does not choose the product's domain, frameworks or architecture.

Evaluation: unit and black-box tests agree on the identity; no template name remains where it
denotes the active project; `make ci` passes. A negative result names the missing owner input
(product name, module path, description) instead of inventing it.

## Capability Registry

Every capability appears exactly once. Status is `planned` while it has open plan tasks and
`shipped` once its current-state page exists and no task remains. Row order is the implementation
line the plan follows.

| # | Capability | Status | How it is evaluated | Implementation |
| --- | --- | --- | --- | --- |
| 1 | `project-foundation` | shipped | `make ci` passes on a fresh clone; planning-tool and black-box CLI tests | [Current](../impl/current/project-foundation.md) |
| 2 | `release-distribution` | shipped | Packaging tests for names, members, reproducibility and refusals; `make dist` in CI | [Current](../impl/current/release-distribution.md) |
| 3 | `project-identity` | planned | Identity tests pass and no stale template identity remains after personalization | [Open work](../impl/plan.md#project-identity----project-identity) |

## Development integrity

[AGENTS.md](../../AGENTS.md) defines the task cycle; the
[planning workflow](../guide/planning-workflow.md) defines task shape, lanes, records and
checkpoints.

- Task records under `docs/impl/records/` keep the full accepted task, amendments, decisions and
  evidence. Current pages describe available behavior and link records. The plan holds only
  unresolved work.
- Tests are deterministic, network-free and build fixtures in `t.TempDir()` or `testing/fstest`.
  Unit and white-box tests sit beside their package; black-box tests and shared test data live in
  [`test/`](../../test/README.md).
- Coverage percentages are diagnostic, never a gate.
- Tests that need a real external system, credentials or human judgment are human-assisted or
  declared runs, never part of `make ci`.

## Success criteria

A team creates a repository from the template, completes the first plan task to give it its own
identity, and then adds capabilities by specifying them here, planning tasks, and closing each
task with a record and a current-state page, while `make ci` stays green. A contributor can tell
what the product promises, what remains, what exists and how each capability is proven from the
repository alone.
