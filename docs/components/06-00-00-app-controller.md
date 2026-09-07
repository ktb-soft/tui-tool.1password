# App controller

`internal/app` is the only `tea.Model` the program runs. It owns focus, the
panes, the overlay, and the terminal size, and it is the only package that
knows both `internal/op` and `internal/ui`.

```go
type Model struct {
	client  op.Client
	vaults  pane.Pane
	items   pane.Pane
	detail  detail.Detail
	overlay tea.Model
	focus   Focus
	status  string
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

`commands.go` wraps `internal/op`, and nothing else in the project does:

```go
func loadVaults(c op.Client) tea.Cmd {
	return func() tea.Msg {
		vaults, err := c.ListVaults()
		if err != nil {
			return opFailedMsg{err}
		}
		return vaultsLoadedMsg(vaults)
	}
}
```

Every wrapper has this shape. `opFailedMsg` sets `status`, which the footer
renders in the error style, and leaves the panes untouched — a failed delete
must not blank the list.

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

`Init` returns `loadVaults`. If the client reports that the user is not signed
in, the program prints the `op` error and exits non-zero rather than showing an
empty UI — `op` handles authentication, and this program never touches
credentials.
