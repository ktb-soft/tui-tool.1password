# App controller

`internal/app` is the only `tea.Model` the program runs. It owns focus, the
panes, the overlay, and the terminal size, and it is the only package that
knows both `internal/op` and `internal/ui`.

```go
type Model struct {
	client  op.Client
	cache   *cache
	vaults  pane.Pane
	items   pane.Pane
	detail  detail.Detail
	overlay tea.Model
	focus   Focus
	width   int
	height  int
}

type Focus int

const (
	VaultPane Focus = iota
	ItemPane
	DetailPane
)
```

## Messages

Every `op` call produces exactly one message. They live in `messages.go`:

```go
type vaultsLoadedMsg []op.Vault
type itemsLoadedMsg  []op.Item
type itemLoadedMsg   op.Item
type writeSucceededMsg struct{ Reload Focus }
type opFailedMsg struct{ Err error }
```

`writeSucceededMsg` carries which pane to refresh, so one handler covers all
six write paths instead of six near-identical ones.

## Commands

`commands.go` wraps `internal/op`, and nothing else in the project does. Each
read goes through the cache, which calls `op` only on a miss:

```go
func loadVaults(client op.Client, store *cache) tea.Cmd {
	return func() tea.Msg {
		vaults, err := store.getVaults(client.ListVaults)
		if err != nil {
			return opFailedMsg{err}
		}
		return vaultsLoadedMsg(vaults)
	}
}
```

Every wrapper has this shape. `opFailedMsg` is turned into a
`list.NewStatusMessage` on the focused pane (`charm-bubbles.md:2202`), which
displays and expires on its own, and leaves the pane's contents untouched — a
failed delete must not blank the list. The app model carries no status field
of its own.

## Cache

`cache.go` holds the vault list, each vault's item list, and each item's field
values, in memory, for the life of the process. `handleWriteSucceeded` drops
what the completed write changed before it issues the reload, keyed on the
`Reload` pane the message already carries, and `R` drops everything. There is
no TTL and nothing reaches disk; see
[ADR 09](../adr/09-00-00-in-memory-cache.md).

## Update

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
```

`Update` dispatches on message type and delegates: sizes go to
`resizePanes`, keys to `handleKey`, load results to the pane they belong to.
Each handler is its own function, so `Update` itself stays a switch and never
approaches the complexity limit.

`handleKey` follows the routing order in
[the keymap](04-00-00-navigation-keymap.md).

## View

```go
func (m Model) View() tea.View
```

Builds the three columns and the footer, then, if `overlay != nil`, places it
over the result. Returns `tea.NewView(s)` with the window title set to the
focused vault — v2 returns `tea.View`, not `string`
(`charm-bubbletea.md:8183`).

## Startup

`Init` returns `loadVaults`. The arriving list loads the vault it highlights,
and that vault's item list loads the item it highlights, so the three panes
fill without any key being pressed. If the client reports that the user is not signed
in, the program prints the `op` error and exits non-zero rather than showing an
empty UI — `op` handles authentication, and this program never touches
credentials.
