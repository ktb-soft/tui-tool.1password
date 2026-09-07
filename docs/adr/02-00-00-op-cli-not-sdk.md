# Shell out to `op`, do not use the Go SDK

## Status

Accepted.

## Context

1Password ships a Go SDK alongside the `op` CLI. The SDK requires the program
to supply a service account token.

## Decision

Shell out to the `op` binary and inherit whatever session the user already
has.

## Consequences

The program holds no credentials, stores no token, and contains no
authentication code. Biometric unlock, desktop app integration, service
accounts, and Connect all work, because `op` implements them and this program
does not know which is in use. `op` missing or not signed in is reported by
`op`, in `op`'s own wording.

The cost is a process per operation — roughly 100–300ms — which is why every
call is a `tea.Cmd` and why selection changes are debounced. It also means the
JSON shape belongs to the CLI and a CLI release could change it; this design
was written against `op 2.36.0`.

The one-way dependency in `internal/op` is what keeps this reversible:
adopting the SDK later replaces the bodies of that package's functions and
touches nothing else.
