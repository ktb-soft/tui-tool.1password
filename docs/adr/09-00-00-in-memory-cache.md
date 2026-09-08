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

A failed read caches nothing, so a retry reaches `op`.
