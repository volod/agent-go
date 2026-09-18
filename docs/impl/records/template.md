# Task Record

Copy to `NNNN-<group>-<task-id>.md` using the
[naming rules](../../guide/planning-workflow.md#record-file-naming), index it in the
[records README](README.md), and replace this paragraph with the task title. Keep evidence
concise and the accepted task text complete.

## Task and scope

- Id / capability / checkpoint: `task-id` / `capability-id` / `checkpoint-id` or none
- State: active, blocked, or accepted (only after every required gate passes).
- Source: plan task or ad hoc request; commit at start and any unrelated dirty files.
- Plan counts at start: open tasks (agent, human) and the next eligible task.
- Accepted task: the full original block, verbatim, in a fenced Markdown block.
- Amendments: none, or each revised block with its reason and who authorized it.

## Implementation

Changed packages and behavior, reused code, decisions and rejected alternatives, compatibility
effects, limitations, and the current-state page link.

## Acceptance evidence

| Gate | Exact command, test or artifact | Result and limit |
| --- | --- | --- |
| Each acceptance gate | Reproducible evidence | pass, fail, not-run, blocked or valid-negative |

Missing, mocked or stale evidence does not prove acceptance.

## Audit handoff

`none identified` with the reviewed scope, or one entry per `AUD-<task-id>-N`: the observation,
blocking or nonblocking, its evidence and impact, and exactly one owner (a plan task or a
checkpoint). A checkpoint also records an invariant-to-evidence table, the disposition of every
incoming note, a refactor verdict, and `proceed`, `proceed-with-nonblocking-notes` or `blocked`.

## Close or resume

Gates passed and remaining, the next action, current-page and index updates, plan counts after,
and capability status changes.
