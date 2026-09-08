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
	revealed bool
	form     *huh.Form
	draft    *form.ItemDraft
	initCmd  tea.Cmd
	focused  bool
	width    int
	height   int
}
```

The detail pane is a `huh` form over the selected item, not a document. It
renders the item's title and one input per field, all in a single `huh.Group`
so the whole item is on screen at once rather than paged. A field that belongs
to a section is titled `section · label`, which is what became of the section
headings the old table drew.

The pane is **blurred by default**: it shows the item but takes no keys until
the item pane hands it focus with `e`. It edits values and the title; the
item's field structure is edited in the overlay form `f` opens. See
[ADR 08](../adr/08-00-00-inline-item-editing.md).

`Focus` and `Blur` both rebuild the form from the stored item, so leaving the
pane discards every edit typed into it — that is the whole implementation of
"esc does not save". `Init` returns the command the rebuilt form needs;
`Completed` reports `huh.StateCompleted`, and `EditedItem` folds the draft back
onto the item the form was seeded from.

A concealed field is built with `EchoMode(huh.EchoModePassword)`, so its value
never reaches the screen. `huh` fixes an input's echo mode at construction, so
`r` toggles `revealed` and rebuilds the form rather than mutating it. Reveal is
per item, not per field, and resets whenever the selected item changes.

`r` and `e` are pressed on the **item** pane, not the detail pane: once the
form has focus every printable key belongs to it.

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

The detail pane has no list, so it renders one line from a `theme` constant
when it is empty: `select an item`, inset by `theme.DetailIndent` so it lines
up with the form fields that replace it rather than sitting against the
border.
