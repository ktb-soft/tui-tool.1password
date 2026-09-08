# In-memory cache in the controller, not in `internal/op`

## Status

Accepted.

## Context

1Password rate-limits requests, and every settled selection change shelled out
to `op` again — moving the item cursor down and back up refetched both items.
Nothing was cached.

The cache had to live somewhere, and there were two places it could go.

## Decision

The cache is `internal/app/cache.go`, held by the root model and read by the
three load commands. `internal/op` is unchanged.

It holds three things, each for the life of the process:

- the vault list
- the item list, per vault ID
- the item with its field values, per vault ID and item ID

The third is filled **in bulk, when a vault is selected**, not per item as the
cursor reaches it. Selecting a vault runs `op item list --vault <id>` and then
one `op item get -` fed the listed ids on stdin, which returns every item with
its fields (`op item get --help`). Both results go into the cache, so every
selection inside that vault is served from memory. A vault of forty items
costs two subprocesses to browse end to end, not forty-one.

An empty vault runs no bulk read: there are no ids to send.

Nothing is written to disk. There is no TTL and no background refresh: entries
are dropped when a write changes them, and `R` drops all of them.

## Alternatives rejected

**A caching `Client` inside `internal/op`.** It would have left `internal/app`
untouched, but it contradicts what `internal/op` is.
[ADR 02](02-00-00-op-cli-not-sdk.md) keeps the CLI swappable for the Go SDK by
making that package nothing but an exec adapter — "adopting the SDK later
replaces the bodies of that package's functions and touches nothing else." A
cache in there is a body that must survive the swap, so the swap stops being a
replacement and becomes a merge. The controller already owns *when* a read
happens — the debounce, the cascade, the reload after a write — and a cache is
a decision about when to read, not about how to run `op`.

**Lazy per-item caching alone.** The first version of this cache fetched an
item the first time the cursor settled on it. That prevents a *repeat* fetch
but not the first pass, and the first pass is the whole cost: reaching the
fortieth item of a vault still ran forty `op item get` calls, which is what
hits the rate limit. Caching without the bulk read solves the smaller half of
the problem.

**A TTL, or a background refresh.** Both generate traffic nobody asked for,
against the rate limit this change exists to relieve, and a TTL still shows a
stale value for as long as it lasts. Invalidating on write is exact, and `R`
covers the only other case: a change made somewhere other than this program.

**Caching inside the `op.Client` value.** `Client` is copied by value
throughout the program; a cache in it would either be copied with it or need a
pointer field, and the type would stop being the plain, comparable
configuration struct it is.

## Consequences

Cached field values are secrets held in process memory. They are never logged,
never written to a file, and never put in an error — the same rule
[ADR 04](04-00-00-secrets-never-in-argv.md) applies to argv, extended to the
cache. That ADR's "the program writes no cache, log, or temp file containing
item data" is about files, and stays true.

A persistent cache is a different decision with a secrets-at-rest problem, and
this one does not settle it.

`writeSucceededMsg.Reload` drives invalidation, so no second notion of what
changed exists. A vault write drops the vault list. An item write drops the
selected vault's item list and every item cached from that vault, which covers
create, edit, and delete without the message having to name a record.

The vault list's per-vault item count goes stale when an item is created or
deleted, until a vault write or `R`. Nothing reloads the vault list on an item
write today, so the count was already stale for the same window before this
change.

A failed **list** caches nothing, so a retry reaches `op`. A failed **bulk
read** is different: the item list still arrived, so the pane fills from it and
the vault falls back to fetching each selection with `GetItem`, exactly as it
behaved before the bulk read existed. A vault the bulk read cannot serve is
slower, never broken, and never empty.

The bulk response is one JSON document carrying every secret in the vault. It
is held in memory only, and the rule the rest of this ADR states applies to it
unchanged: it is never logged, never written to a file, and never put in an
error message. Only ids go to the subprocess, and they go on stdin, so
[ADR 04](04-00-00-secrets-never-in-argv.md) holds for the read path too.

The bulk read is slower than a single item, so the item pane runs
`list.StartSpinner` while it is in flight and stops it when the list arrives or
the read fails.
