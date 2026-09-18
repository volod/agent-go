# Capability name

<!-- Copy to docs/openspec/<capability-id>.md when a capability outgrows its section in spec.md,
link the page from that section, replace every <placeholder> and delete this comment. -->

- Capability: `<capability-id>` in the [registry](spec.md#capability-registry)
- Problem: what an operator or user cannot do or trust today, in their terms.

## Behavior

Commands, flags and defaults; inputs and outputs; file formats; error cases and exit codes. Mark
contracts that other capabilities consume.

## Boundaries

What this capability does not do, and which capability owns the adjacent behavior.

## Design

Packages and import edges from the [architecture](architecture.md); data flow, state,
concurrency and cancellation; new dependencies from the
[dependency table](spec.md#dependencies).

## Evaluation

| Gate | Evidence |
| --- | --- |
| Acceptance signal | Deterministic test, fixture or declared run that proves it |

Valid negative result: what a supported "no" looks like and what it leaves in place.
