# Project rules

Canonical rules for every agent and contributor. `CLAUDE.md`, `GEMINI.md` and `.cursor/rules/`
only point here; never put rules in them. Read the selected task, the code it touches and the
specification sections it links. Load other guidance only when a condition under
[Read when needed](#read-when-needed) applies.

## Project

`agent-go` is a template for Go command-line projects built by people and coding agents. The
[specification](docs/openspec/spec.md) says what the product must do and how each capability is
evaluated, the [plan](docs/impl/plan.md) holds only remaining work, [task records](docs/impl/records/README.md)
keep evidence and decisions, and [current state](docs/impl/current.md) describes what exists.
The [architecture](docs/openspec/architecture.md) owns the package layout and dependency direction.

## Guardrails

- Preserve unrelated work. Diagnostic or review requests do not authorize code changes.
- Do not commit, push, rewrite history or revert user changes unless explicitly asked.
- Go 1.27+, module `github.com/volod/agent-go`. Use the Make targets; each runs a plain `go`
  command listed in the [development guide](docs/guide/development.md#make-targets).
- Prefer the standard library. Add a dependency only after adding it to the
  [dependency table](docs/openspec/spec.md#dependencies); commit `go.mod` and `go.sum` together.
- Never hardcode machine-specific paths. Never put secrets in code, logs, fixtures or docs.
- Use ASCII in code, logs and docs.

## Go conventions

- `cmd/<name>/main.go` only wires the process (signal context, streams, `os.Exit`) and calls
  `cli.Run`. Product code lives in `internal/`, dev-only commands in `tools/`, black-box tests
  and shared test data in `test/`. Add `pkg/` only for code another module imports.
- Name packages for what they provide; no `util`, `common` or `helpers`. Follow the
  [dependency direction](docs/openspec/architecture.md#dependency-direction); put platform code
  in build-tagged files (`_linux.go`, `_windows.go`).
- Only `internal/cli` reads flags and the environment, writes to stdout and stderr, and maps
  errors to exit codes. Other packages take typed configuration, return errors and never call
  `os.Exit`, `log.Fatal` or `panic` for expected failures.
- Wrap errors with context (`fmt.Errorf("load plan: %w", err)`) and test them with `errors.Is`
  or `errors.As`. Handle every error once: return it or log it, not both.
- Functions that do I/O or may block take `ctx context.Context` first and stop when it is
  canceled. Never store a context in a struct. Every goroutine has an owner that waits for it.
- Accept interfaces, return concrete types; declare an interface where it is consumed. Avoid
  mutable package-level state; pass loggers (`*slog.Logger`), clocks and filesystems (`fs.FS`).
- Tests are table-driven with `t.Run`, deterministic and network-free. They build fixtures in
  `t.TempDir()` or `testing/fstest`, use `t.Context()`, and assert behavior, not incidental
  implementation details. A bug fix starts with a failing regression test.
- Keep files at about 300 lines or less; split at real seams.

## Task cycle

1. Run `make plan-status`, select one eligible task and note the counts. Check its dependencies
   and the records they link. Do not start blocked work.
2. Create the task record from the [template](docs/impl/records/template.md) using the
   [naming rules](docs/guide/planning-workflow.md#record-file-naming), paste the full task text
   and index it. Identify the affected packages and reusable code; do not silently broaden scope.
3. Implement and self-review. Tests cover the happy path, the main edge cases and a regression
   for every bug fixed.
4. Verify with the relevant tests and `make ci`. Record failures and unrun checks honestly. Fix
   causes; never weaken a gate. Coverage is diagnostic, never a gate. Failed acceptance keeps the
   task open.
5. Before stopping, update the record with evidence, decisions, audit notes and the next action.
   On acceptance: update the narrow `docs/impl/current/` page and link the record; remove the task
   from the plan and replace references to its id with the record link; mark the capability
   `shipped` when its last task is done. Run `make lint-spec-plan lint-doc-links`, report the
   task counts before and after and the next eligible task, inspect `git status` and remove
   temporary files.

## Read when needed

- **Adding or changing capabilities or tasks:** the
  [planning workflow](docs/guide/planning-workflow.md). Specify behavior and evaluation before
  planning or coding.
- **Concerns outside the task, or a checkpoint task:** the
  [audit rules](docs/guide/planning-workflow.md#audit-notes-and-checkpoints). Route each concern
  to exactly one owner.
- **Human-assisted tasks:** agents prepare inputs and report what is needed; they never mark a
  human task done. See [task lanes](docs/guide/planning-workflow.md#task-lanes).
- **Toolchain, Make targets, CI or test layout:** the
  [development guide](docs/guide/development.md) and the [test layout](test/README.md).
