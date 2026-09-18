# agent-go Implementation Plan

Forward-only: this file holds only work that remains. Product behavior, boundaries and evaluation
belong in the [specification](../openspec/spec.md); task shape, statuses, ordering and records in
the [planning workflow](../guide/planning-workflow.md); available behavior in
[current state](current.md). Run `make plan-status` for counts and the next eligible task.

Every task keeps `make ci` green, adds deterministic, network-free tests, and adds each new
dependency to the [dependency table](../openspec/spec.md#dependencies) in the same change.

## Agent Implementation Tasks

### Project identity -- `project-identity`

#### personalize-template-project

Give a repository created from this template its own identity before any product work starts.

- Serves: `project-identity` -- [Project identity](../openspec/spec.md#project-identity)
- Agent status: CLEAR
- Dependencies: none.
- User-visible outcome: The module path, command, README and specification name and describe the
  new product; `<name> version` prints the new name and module path.
- Scope boundary: Rename the module (`go mod edit -module`), `cmd/agent-go/`, `buildinfo.Name`,
  `APP` in `make/config.mk`, imports and test expectations. Rewrite the README and the purpose,
  scope and identity sections of the specification. Keep the planning tooling, gates and
  documentation lifecycle. Do not add capabilities, dependencies or architecture.
- Data and artifact paths: `go.mod`, `cmd/`, `internal/`, `test/`, `make/config.mk`, `README.md`,
  `AGENTS.md`, `docs/`.
- Execution path: Take the product name, module path and one-line description from the owner, or
  derive the module path from `git remote get-url origin`. Apply the renames, then run
  `make ci`, `make build` and `bin/<name> version`.
- Acceptance gates: `git grep -n agent-go` finds no reference that denotes the active project;
  `make ci` passes; `bin/<name> version` prints the new name and module path; `project-identity`
  is `shipped` with a current-state page. A negative result names the missing owner input
  instead of inventing it.
- Documentation target: new `docs/impl/current/project-identity.md`, linked from
  [current state](current.md) and the registry.
- Review checkpoint: none.

## Human-Assisted Tasks

None open. Add a task here when acceptance needs human judgment, authorization, private access or
spending authority.
