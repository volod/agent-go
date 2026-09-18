# Planning Workflow

[AGENTS.md](../../AGENTS.md) gives the task cycle. Read only the section you need here.
`make lint-spec-plan` parses the formats below, so keep headings and `- Field name:` prefixes
exact.

## Capability changes

A need discovered during work is classified once: a chore (do it now or drop it), part of the
current task's self-review, more work for a registered capability (a new task), or a new
capability. A new or misshaped capability goes through these steps before any plan task or code:

1. State the operator or user problem, not an implementation preference.
2. Amend the owning section of the [specification](../openspec/spec.md), or a page copied from
   the [page template](../openspec/template.md): behavior, what it does not do, evaluation, the
   acceptance signal and a valid negative result.
3. Add a `planned` registry row in implementation order and a row to the
   [record group table](#record-file-naming).
4. Add tasks under the capability's group in both lanes as needed.
5. When its last task is accepted, mark the row `shipped` and link its current-state page.

## Task shape

Agent tasks use a stable lowercase slug under their capability group:

```markdown
### Capability name -- `capability-id`

#### stable-task-id

State the unresolved problem in one or two sentences.

- Serves: `capability-id` -- [Owning spec section](../openspec/spec.md#section)
- Agent status: CLEAR
- Dependencies: Open task ids in backticks, or accepted record links; `none.` when there are none.
- User-visible outcome: What becomes possible or trustworthy.
- Scope boundary: Included and excluded behavior.
- Data and artifact paths: Repository-relative paths and outputs.
- Execution path: Packages, fixtures, commands and declared runs.
- Acceptance gates: Checks, the acceptance signal and the valid negative result.
- Documentation target: The narrow current-state page.
- Review checkpoint: The owning checkpoint task id, or `none.`
```

Optional fields go after `Agent status`: `Task kind: refactor | checkpoint` and `Research: yes`
(a supported negative result is acceptable; it does not replace acceptance gates). A heading may
end with `(optional)` for refinements, which follow required tasks in their group. Field values
may continue on indented lines.

Human tasks use:

```markdown
#### stable-task-id

- Serves: `capability-id` -- [Owning spec section](../openspec/spec.md#section)
- Human status: HUMAN-GATED
- Dependencies: Open task ids or record links, or `none.`
- Requested input or decision: What the human provides or decides, and how to check it.
- Unblocks: Backticked ids of the tasks that wait for it.
```

## Task lanes

| Lane | Status | Acceptance needs |
| --- | --- | --- |
| Agent Implementation Tasks | `CLEAR` | Local code, tests and docs |
| Agent Implementation Tasks | `RUN NEEDED` | A declared heavier run (network, live service, another OS) |
| Human-Assisted Tasks | `BLOCKED BY HUMAN` | Human-provided input or access (credentials, data) |
| Human-Assisted Tasks | `HUMAN-GATED` | Human judgment, authorization or spending decision |

An agent task that depends on a human task stays ineligible until the decision is recorded.
Agents never approve a human task themselves.

## Ordering

Capability groups follow [registry](../openspec/spec.md#capability-registry) order in both lanes.
Within a group: prerequisites first, work that changes later inputs next, cheap deterministic work
before expensive runs, optional refinements and checkpoints last. Change priority by reordering
the registry, never by task wording. Dependencies must resolve to open tasks or existing records,
without cycles; `make plan-status` reports the next eligible task and what else may run in
parallel.

## Durable task records

At task start, copy the [record template](../impl/records/template.md) into a sequenced file and
index it in the [records README](../impl/records/README.md). The record keeps the full original
task and any amendments after the task leaves the plan, plus evidence and decisions. Current pages
describe available behavior and link records. Future work goes to the plan with one owner.

### Record file naming

Filenames are `NNNN-<group>-<task-id>.md`:

1. **Sequence:** four digits, one more than the highest existing record; never reused.
2. **Group:** the abbreviation of the owning capability from the table below.
3. **Task id:** the unchanged plan slug; the record's first scope line names it.

| Capability id | Group abbrev |
| --- | --- |
| `project-foundation` | `foundation` |
| `release-distribution` | `dist` |
| `project-identity` | `identity` |
| `governance` (agent rules, workflow, audits) | `govern` |

## Audit notes and checkpoints

Self-review every task and record `none identified` or notes `AUD-<task-id>-N` with evidence.
Route a concern outside the task's scope to exactly one owner: a blocking repair task placed
before dependent work, or the named checkpoint for nonblocking concerns. Do not broaden the
current task.

A checkpoint task reviews the records, code, tests and notes routed to it. It records an
invariant-to-evidence table, each note's disposition, a refactor verdict, and `proceed`,
`proceed-with-nonblocking-notes` or `blocked`. A blocker keeps the checkpoint open until its
repair passes.

## Completion transition

Before removing a task from the plan, map every acceptance gate to evidence in the record, route
every note, link the record from the current-state page, replace references to the task id in
the plan with the record link, and update the registry status when the capability is complete.
Never put completion notes, dates, measurements or history in `plan.md`. Failed acceptance leaves
the task open, with partial results and the next action in its record.
