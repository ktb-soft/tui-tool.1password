# Bulk vault fetch

The in-memory cache ([ADR 09](../adr/09-00-00-in-memory-cache.md)) stopped
repeat fetches but not the first pass. Moving the selector down a vault still
ran one `op item get` per row the cursor settled on, so reaching the fortieth
item cost forty requests and hit the 1Password rate limit anyway.

Selecting a vault now reads every item's fields in one call.
`op item get -` takes a JSON array of object specifiers on stdin and returns an
item for each one carrying an `id` (`op item get --help`), so the vault costs
one `op item list` plus one `op item get -`, and the cache is seeded for every
item in it before the cursor moves.

## Measured

Counted subprocess invocations for browsing a vault of 40 items — the list load
plus a settled selection on every row. Measured with the counting fake `op` in
`internal/app/bulk_fetch_test.go`, which appends a line per exec.

| | subprocesses |
|---|---|
| before (lazy per-item cache) | 41 |
| after (bulk read) | 2 |

The 41 was measured by mutating `fetchVaultItems` back to the plain
`ListItems` call and rerunning `TestBrowsingAWholeVaultCostsTwoSubprocesses`.

Other counts changed with it. A vault selection is now two calls rather than
one, so the cache tests that count a single item-list read expect two, and a
refresh that reloads vaults, one item list, and one item's fields is three
rather than four — the item's fields arrive with the bulk read.

## Failure

A failed bulk read is not fatal. The item list already arrived, so the pane
fills from it, nothing is cached for that vault, and each selection falls back
to `GetItem` — the behavior the program had before this change. An empty vault
sends no ids and runs no bulk call at all.

## Loading feedback

The bulk read is slower than a single item, so the item pane starts
`list.StartSpinner` when a vault selection settles and stops it when the list
lands or the read fails. `spinner.TickMsg` is routed to the item pane
explicitly, since the vault pane holds focus while the read runs and the
message would otherwise never reach the pane that is loading.
`docs/components/03-00-00-pane-view.md` described this spinner already; until
now nothing called it.
