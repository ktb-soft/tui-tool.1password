# Vertical-slice build order

## Status

Accepted.

## Context

A layered structure — model, view, controller — invites building layer by
layer: all the `op` calls, then all the panes, then the wiring. That defers
every integration failure to the end, when the most code is in flight and the
least of it has been run.

## Decision

Build one vertical slice at a time. A slice is the data, the view, and the
controller wiring for a single pane, working end to end in a running program
before the next slice starts.

1. **Vault pane** — `ListVaults`, `Pane`, focus, the app skeleton
2. **Item pane** — `ListItems`, the second `Pane`, selection cascade, debounce
3. **Detail pane** — `GetItem`, `Detail`, section grouping, reveal
4. **Vault CRUD** — create, rename, delete, confirmation
5. **Item CRUD** — create and edit forms, stdin templates, delete
6. **Field CRUD** — the field editor inside the item form
7. **Polish** — help overlay, filter, clipboard, resize edge cases

## Consequences

Slice 1 produces a program that runs and does something real. Every later
slice extends a working program, so a break is attributable to the change that
caused it.

The layer boundaries still hold — a slice touches all three packages, it does
not collapse them. `internal/op` accumulates one function per slice instead of
being written whole up front, which keeps it to functions something actually
calls.

The cost is that shared code appears in the second slice that needs it, not
the first: `Pane` is written for vaults and generalized when items arrive.
That is the intended order — the abstraction shows up once there are two cases
to abstract over.
