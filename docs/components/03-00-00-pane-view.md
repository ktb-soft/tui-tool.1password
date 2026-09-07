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
the terminal width by each ratio, subtracts the border width, and calls
`SetSize` on each pane. The last column absorbs the rounding remainder so the
three always sum to the terminal width exactly.

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

Fields render as an aligned two-column list, grouped by section. A field where
`IsSecret()` is true and `revealed[field.ID]` is false renders as `••••••••`.
Reveal is per field and resets whenever the selected item changes — nothing
stays revealed after navigating away.

The detail pane is a `viewport` rather than a list: fields are read, scrolled,
and edited as a whole item, not selected one at a time. Field-level editing
happens in the edit form, which shows every field at once.

## Empty and loading states

Each pane renders one centered line when it has nothing: `no vaults`,
`select a vault`, `select an item`, `loading…`. That string is a `theme`
constant per pane, so the phrasing is defined once.
