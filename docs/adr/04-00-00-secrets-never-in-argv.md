# Secrets never in argv

## Status

Accepted.

## Context

`op item create` and `op item edit` accept assignment statements as command
arguments — `password=hunter2`. The `op` help says so plainly:

> Caution: Command arguments can be visible to other processes on your
> machine.

Argv is readable by other local processes, so any of them can read a password
out of `ps` while the command runs.

Both commands also accept an item JSON template on standard input, with `-` as
the first argument (`op item create --help`, `op item edit --help`).

## Decision

Every write goes through stdin as a JSON template. `internal/op` never builds
an assignment statement containing a value.

```
op item create --vault <id> -
op item edit <id> -
```

Non-sensitive arguments — vault IDs, item IDs, category, `--name` on a vault —
stay in argv.

## Consequences

`Client.run` takes a `stdin []byte` parameter on every call, used or not, so
there is no second write path that could skip it.

Reads are constrained the same way. `op item get --format json` returns
concealed values in plaintext on stdout; those values live only in memory — in the
`op.Item` the detail pane holds and in the in-memory cache behind it
([ADR 09](09-00-00-in-memory-cache.md)) — are masked on screen unless
explicitly revealed, and reveal state is dropped when the selection changes.
The program writes no cache file, log, or temp file containing item data — which also rules out
the `--template=<file>` form of these commands, since it would put a plaintext
secret on disk.
