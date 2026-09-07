# Parallel slice construction

## Status

Accepted.

## Context

[ADR 05](05-00-00-vertical-slice-build-order.md) orders the work as seven
sequential vertical slices. Several of those slices do not actually depend on
each other, and building them one at a time leaves work on the table when more
than one agent is available.

The obstacle to parallelism is not the slice boundaries — those are already
clean — it is that every slice touches `internal/app`. Three agents editing
`app/update.go` at once produces three-way merge conflicts in the one file
where a bad merge is hardest to spot.

## Decision

Freeze a contract first, partition by **file** rather than by package, and run
the independent slices in two waves.

### Slice 0 — the spine, built serially, by one agent

Nothing runs in parallel until this exists and compiles:

- `internal/op` — every type (`Vault`, `Item`, `Field`) and every function
  **signature**, with stub bodies returning `errNotImplemented`, plus
  `Client.run` fully implemented.
- `internal/ui/theme` — every color, dimension, ratio, and string constant.
- `internal/ui/keymap` — the whole `KeyMap` struct and all bindings.
- `internal/app/messages.go` — every message type.
- `internal/app/app.go` and `update.go` — the root model and the dispatch
  switch, which already calls per-vertical handlers that do nothing yet.
- `internal/app/vaults.go`, `items.go`, `detail.go` — those handlers, as
  no-op stubs, plus `commands.go` for the shared debounce.
- `internal/ui/pane/pane.go` and `internal/ui/detail/detail.go` — skeletons.
  `app.go` holds a `pane.Pane` and a `detail.Detail`, so the types have to
  exist in slice 0 or every wave-1 agent adds the same field to `app.go` and
  collides there. Those files pass to their owning verticals afterwards.
- `testdata/` — real `op --format=json` output for vaults, an item list, and
  several item categories, recorded once and scrubbed of live secrets.

At the end of slice 0 the program runs and shows three empty bordered panes.
This is the contract. Every later agent codes against it and none of them
changes it.

### File partition

Each vertical owns its own files, in every package it touches:

| Vertical | Owns |
|---|---|
| Vaults | `op/vault.go`, `ui/form/vault.go`, `app/vaults.go` |
| Items | `op/item.go`, `ui/form/item.go`, `app/items.go` |
| Detail | `ui/detail/*.go`, `app/detail.go` |
| Panes | `ui/pane/*.go` |

`app/update.go` dispatches into `handleVaultKey`, `updateItems`, and so on —
written in slice 0, unchanged afterwards. No two agents open the same file, so
merges are additive.

Two handler shapes, and only two:

```go
func (m Model) handleVaultKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool)
func (m Model) updateVaults(msg tea.Msg) (tea.Model, tea.Cmd)
```

The `bool` reports whether the key was consumed; when it is false `update.go`
forwards the key to the focused component, which is what gives `list` its own
movement and filter keys for free. `updateX` receives every message routed to
that vertical — its load message, `selectionChangedMsg` for its pane, and
`writeSucceededMsg` naming its pane — and must type-switch on `msg` itself.

### Wave 1 — three agents, read paths only

Vault pane, item pane, and detail pane. All three depend only on slice 0 and
on nothing from each other. Each works against `testdata/`, so the detail
agent is not blocked waiting for the item agent to finish loading real items.

Merge, then integrate serially: the selection cascade between panes is the one
piece that is genuinely cross-vertical, and it is written once, by one agent,
after all three land.

### Wave 2 — three agents, write paths

Vault CRUD, item CRUD, field CRUD. Each extends the vertical it belongs to,
in the files that vertical already owns.

### Then polish, serially

Help overlay, filter tuning, clipboard, resize edge cases. These cut across
everything and are small; parallelizing them costs more than it saves.

### Mechanics

Each agent works in its own git worktree on its own branch, per the global
standard, and merges by PR. An agent's slice is done when the program builds,
its tests pass, and it runs — not when its files exist.

## What running it actually taught

Both waves ran. Three amendments, from
[wave 1](../log/03-00-00-wave-1-integration.md) and
[wave 2](../log/04-00-00-wave-2-integration.md):

1. **Give every agent its own file in each shared package**, not just its own
   package. `theme` would have been a three-way collision otherwise; each agent
   got `theme/<vertical>.go`. The one collision that did happen — two `drain`
   test helpers in the `form` package — was in the one package where two agents
   were told to create files at the same time.
2. **A slice is only as frozen as it is exercised.** Slice 0 met its "it runs"
   bar with a program that had no forms in it, and froze an overlay seam that
   could never complete one. If a seam has no caller yet, its contract is a
   guess. Prefer freezing seams that at least one vertical uses in the same
   slice.
3. **A green test on a branch can encode the absence of a feature another
   branch is adding.** One wave 2 test asserted no form opens on the vault
   pane, which stopped being true when vault CRUD merged. Expect a small number
   of these per wave and treat them as integration work, not as a regression.

The count is a consequence of the partition, not a target: wave 2 was two
agents rather than three because field editing lives inside the item form and
could not be given its own owner.

## Consequences

Wall-clock time for the read paths drops to roughly one slice instead of
three, and the same again for the write paths.

The costs are real and worth stating:

- **The contract has to be right up front.** This is the opposite of ADR 05's
  "the abstraction shows up when there are two cases to abstract over."
  Slice 0 commits to `Pane`, to the message set, and to the `op` signatures
  before any of them has a second caller. If a wave-1 agent finds the contract
  wrong, the correct response is to stop the wave, fix the contract serially,
  and restart — not to have three agents patch around it independently.
- **Attributability weakens.** ADR 05's main benefit is that a break traces to
  one change. Three merged branches dilute that. The file partition and the
  per-slice "it runs" bar are what keep it from disappearing entirely.
- **Coordination is not free.** Three agents plus a merge and an integration
  pass is worth it for three substantial slices. It is not worth it for two
  small ones — run those serially.

Rejected: partitioning by layer, with one agent per package. Every agent would
then be blocked on every other, which is the layer-by-layer build ADR 05
exists to prevent.
