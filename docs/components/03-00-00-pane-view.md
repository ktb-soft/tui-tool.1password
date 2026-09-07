# Pane view

`internal/ui` renders. It never calls `op` and never decides what data to
fetch.

## Layout

Three columns joined horizontally, one footer below:

```go
columns := lipgloss.JoinHorizontal(lipgloss.Top,
	vaults.View(), items.View(), detail.View())
return lipgloss.JoinVertical(lipgloss.Left, columns, help.View(keys))
```

Widths come from one place, `internal/ui/theme`:

```go
const (
	VaultPaneRatio  = 0.20
	ItemPaneRatio   = 0.35
	DetailPaneRatio = 0.45
	FooterHeight    = 1
)
```

On `tea.WindowSizeMsg` (`charm-bubbletea.md:11592`) the controller multiplies
the terminal width by each ratio and calls `SetSize` on each pane with the
full outer width. The last column absorbs the rounding remainder so the three
always sum to the terminal width exactly.

A lipgloss `Style.Width` is the *outer* width, border included, so `Pane.View`
passes the pane width straight through and only `list.SetSize` subtracts
`theme.BorderWidth` for the inner content.

`help.Model` renders its full short help regardless of the width it was given,
so the footer is capped with `theme.Footer(width)`, which is a
`lipgloss.MaxWidth`. Without it the footer is wider than the panes and
`JoinVertical` pads every row out to match.

## Pane

```go
type Pane struct {
	Title   string
	list    list.Model
	focused bool
	width   int
	height  int
}
```

The vault and item panes are both `Pane`; only their titles and contents
differ. `Pane.View()` renders the list inside a border whose style depends on
`focused` — the pattern the Charm examples use for multi-model focus
(`charm-bubbletea.md:1411`, `:6046`):

```go
var (
	FocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Accent)
	BlurredBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Muted)
)
```

Both borders are `RoundedBorder`, not one rounded and one hidden, so the
layout does not shift by a cell when focus moves.

Every color and border in the program is a constant in `theme`. No package
outside `theme` calls `lipgloss.Color`.

## Detail

```go
type Detail struct {
	item     op.Item
	revealed map[string]bool
	viewport viewport.Model
}
```

Fields render through `lipgloss/v2/table`, one table per section, joined
vertically inside the viewport. The table computes its own column widths
(`charm-lipgloss.md:4064`, `:4112`), so no code in this project measures a
string or pads a cell.

```go
table.New().
	Border(lipgloss.HiddenBorder()).
	StyleFunc(theme.FieldCell).
	Rows(rows...)
```

A field where
`IsSecret()` is true and `revealed[field.ID]` is false renders as `••••••••`.
Reveal is per field and resets whenever the selected item changes — nothing
stays revealed after navigating away.

The detail pane is a `viewport` rather than a list: fields are read, scrolled,
and edited as a whole item, not selected one at a time. Field-level editing
happens in the edit form, which shows every field at once.

## Empty, loading, and status states

`list.Model` already provides all three, so `Pane` configures them rather than
rendering them:

- **Empty** — `SetStatusBarItemName("vault", "vaults")`
  (`charm-bubbles.md:1931`) gives the list its own empty and count text.
- **Loading** — `StartSpinner()` / `StopSpinner()`
  (`charm-bubbles.md:2185`) while an `op` call for that pane is in flight.
  `Pane` exposes them; the controller calls them when it dispatches and when
  the result lands.
- **Transient status** — `NewStatusMessage()` (`charm-bubbles.md:2202`) for
  "deleted", "copied", and `op` errors, which expires on its own.

The detail pane has no list, so it renders one centered line from a `theme`
constant when it is empty: `select an item`.
