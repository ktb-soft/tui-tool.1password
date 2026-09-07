# op client

`internal/op` is the only package in the project that runs a subprocess. Every
other package receives Go values.

## Client

```go
type Client struct {
	Path    string
	Account string
}

func (c Client) run(args []string, stdin []byte) ([]byte, error)
```

`run` is the single choke point: it builds `exec.Command(c.Path, args...)`,
appends `--format=json` and `--account`, pipes `stdin`, captures stdout and
stderr separately, and turns a non-zero exit into an `OpError` carrying the
stderr text. Every exported function in the package is a thin wrapper over
`run`, so the exec flags, the account handling, and the error shape are each
written once.

```go
type OpError struct {
	Args   []string
	Stderr string
	Err    error
}
```

`Error()` renders `op vault delete: <stderr>`. The controller shows that text
verbatim in the status bar — `op` already writes actionable messages, and
rewriting them loses detail.

## Commands used

| Operation | Command |
|---|---|
| List vaults | `op vault list` |
| Create vault | `op vault create <name>` |
| Rename vault | `op vault edit <id> --name <name>` |
| Delete vault | `op vault delete <id>` |
| List items | `op item list --vault <id>` |
| Get item | `op item get <id> --vault <id>` |
| Create item | `op item create -` (JSON template on stdin) |
| Edit item | `op item edit <id> -` (JSON template on stdin) |
| Delete item | `op item delete <id> --vault <id>` |

Writes go through stdin, never assignment statements on the command line — see
[the ADR](../adr/04-00-00-secrets-never-in-argv.md). `op item create` and
`op item edit` both accept a piped JSON template with `-` as the first
argument (`op item create --help`, `op item edit --help`).

## Concurrency

Every call is slow enough to block a frame, so nothing in this package is
called from `Update`. The controller wraps each function in a `tea.Cmd`; see
[the app controller](06-00-00-app-controller.md).

Only one write is ever in flight. The controller blocks input behind a form
overlay while a write runs, so no queue or mutex is needed.

## Testing

`Client.Path` makes the binary injectable. Tests point it at a fixture script
that echoes recorded `op` JSON and asserts the argv it was handed. Nothing
mocks the parsing or the domain types — those are pure and tested directly.
