# agent-go

A copy-ready Go project skeleton for teams that want coding agents to work from one set of rules,
one product specification and one forward plan.

A fresh copy builds a static `agent-go` command, passes its quality gates, and has a
machine-checked plan whose first task gives the new repository its own identity. The Go layout,
gates and planning workflow are meant to stay; the name and the product are yours.

## Quick start

Requirements: Go 1.27+, GNU Make and Git.

```bash
make ci                  # format, vet, staticcheck, tidy, tests, build, plan and link checks
make build               # static binary in bin/
bin/agent-go version     # agent-go v0.0.0-... (github.com/volod/agent-go, go1.27.1, linux/amd64)
make plan-status         # next task: personalize-template-project
```

## Start a project from the template

1. Create a repository from this template and clone it.
2. Ask your agent to take the next task from `make plan-status`. The first one,
   `personalize-template-project`, renames the module, command and docs; give it the product
   name, module path and a one-line description.
3. Describe the product's first capability in the [specification](docs/openspec/spec.md), then
   plan and implement it as the [planning workflow](docs/guide/planning-workflow.md) describes.

## Daily commands

| Command | Purpose |
| --- | --- |
| `make help` | List every target |
| `make run ARGS="..."` | Run the command from source |
| `make test` | Run all tests |
| `make ci` | Run the required gate (what CI runs) |
| `make build-all` | Cross-build for Linux, macOS and Windows |
| `make dist` | Package release archives and `SHA256SUMS` into `dist/` |
| `make plan-status` | Count open tasks and show the next eligible one |

The [development guide](docs/guide/development.md) lists every target and the `go` command behind
it.

## Releases

Push a version tag; the release workflow checks, builds and publishes reproducible archives for
every platform with `SHA256SUMS`:

```bash
git tag v0.1.0 && git push origin v0.1.0
```

## How work flows

```text
docs/openspec/spec.md     what the product must do and how each capability is evaluated
        |
docs/impl/plan.md         only work that remains, ordered by the capability registry
        |
docs/impl/records/        one record per task: full scope, decisions, evidence
        |
docs/impl/current.md      what exists now, linking the records that prove it
```

`make lint-spec-plan` fails when these documents disagree, and `make lint-doc-links` fails on a
broken relative link or anchor.

## Layout

```text
cmd/agent-go/       main: signal context, streams, exit code
internal/cli/       command-line contract: parsing, usage, logger, exit codes
internal/buildinfo/ build identity from the Go toolchain's VCS stamp
tools/plancheck/    dev-only planning and link checks
tools/dist/         dev-only release packaging
test/               black-box tests and shared test data
make/               Makefile fragments
docs/               specification, plan, records, current state, guides
```

## Agent support

[AGENTS.md](AGENTS.md) is the only rule file. `CLAUDE.md` and `GEMINI.md` import it, and the
Cursor rule in `.cursor/rules/` points to it; Codex and other agents read `AGENTS.md` directly.

## License

MIT. See [LICENSE](LICENSE).
