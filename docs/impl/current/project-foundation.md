# Project Foundation

## Layout and command

Module `github.com/volod/agent-go` on Go 1.27 with no runtime dependencies. The
[architecture](../../openspec/architecture.md) lists the tree and the dependency direction.

- `cmd/agent-go/main.go` creates a context canceled by SIGINT or SIGTERM, calls `cli.Run` and
  exits with its code.
- `internal/cli` holds the command table (`help`, `version`), usage text, the stderr
  `log/slog` logger and the exit-code mapping of the
  [command-line contract](../../openspec/spec.md#command-line-contract).
- `internal/buildinfo` reads the identity from `runtime/debug.ReadBuildInfo`: the version is the
  Go toolchain's VCS stamp, or `devel` under `go run` and `go test`.

## Gates

`make ci` runs `fmt-check` (`gofmt -s`), `vet`, `lint` (staticcheck pinned by the `tool`
directive in `go.mod`), `tidy-check` (`go mod tidy -diff`), `test`, `build-all` (every
`PLATFORMS` entry), `lint-spec-plan` and `lint-doc-links`. `.github/workflows/ci.yml` runs
`make ci`, `make test-race` and `make dist` on `ubuntu-latest` for pushes to `main` and for pull
requests. The [development guide](../../guide/development.md) lists every target.

## Planning tooling

`tools/plancheck` is a dev-only command behind `make lint-spec-plan`, `make lint-doc-links` and
`make plan-status`. It checks that registry rows have valid statuses, shipped rows link a
current page, plan groups follow registry order, tasks carry every field of their lane, dependencies
resolve to open tasks or existing records without cycles, and records are named, indexed and
unique. It checks relative Markdown links and heading anchors outside fenced code, and reports
the next eligible task.

## Agent rules

`AGENTS.md` is the only rule source. `CLAUDE.md` and `GEMINI.md` import it with `@`, and
`.cursor/rules/project-rules.mdc` links it, so rules change in one place.

## Tests

- `internal/cli/cli_test.go`: commands, usage, stream separation and exit codes.
- `internal/buildinfo/buildinfo_test.go`: version and module selection from build information.
- `test/integration/cli_test.go`: builds the binary and checks its output and exit codes.
- `tools/plancheck/plancheck_test.go`: each lint defect on `testing/fstest` fixtures, link and
  anchor checks, status and command exit codes.
