# Current Implementation

This index describes behavior available now. Product intent is in the
[specification](../openspec/spec.md), remaining work in the [plan](plan.md), and evidence and
decisions in the [task records](records/README.md).

## Documentation shape

Each capability gets one page under `current/` when its first task is accepted, linked from the
table below and from its registry row. A page states behavior, packages, commands and tests, and
links the records that prove them. When a page grows, it becomes a short index of focused pages
under `current/<capability>/`. Plans, dates and history do not belong here.

## Areas

| Area | Owns | State |
| --- | --- | --- |
| [Project foundation](current/project-foundation.md) | Layout, `agent-go` command seam, Make targets, quality gates, CI, agent rules, planning tooling | Shipped |
| [Release distribution](current/release-distribution.md) | `make dist`, `tools/dist`, reproducible archives, checksums, tag release workflow | Shipped |
