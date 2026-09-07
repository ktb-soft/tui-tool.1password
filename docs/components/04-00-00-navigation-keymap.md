# Navigation and keymap

Every binding in the program is a field of one struct in
`internal/ui/keymap`. The footer help and the `?` overlay are generated from
that struct, so a binding and its documentation cannot drift apart.

```go
type KeyMap struct {
	Up, Down       key.Binding
	PaneLeft       key.Binding
	PaneRight      key.Binding
	Create, Edit   key.Binding
	Delete         key.Binding
	Reveal, Copy   key.Binding
	Filter, Help   key.Binding
	Confirm, Cancel key.Binding
	Quit           key.Binding
}
```

Arrow keys and vim keys are alternates of one binding, not separate bindings:

```go
Up: key.NewBinding(
	key.WithKeys("up", "k"),
	key.WithHelp("↑/k", "up"),
),
PaneLeft: key.NewBinding(
	key.WithKeys("left", "h"),
	key.WithHelp("←/h", "pane left"),
),
```

`KeyMap` implements `help.KeyMap` — `ShortHelp()` returns the footer row,
`FullHelp()` the columns for the `?` overlay (`charm-bubbles.md:888`).

## Movement

`j`/`k` and `↑`/`↓` move within the focused pane. `h`/`l` and `←`/`→` move
focus between panes. Focus is a single `int` on the app model, clamped to
`[0, 2]`; see [the ADR](../adr/03-00-00-column-focus-navigation.md).

Moving right from the vault pane is only allowed once items have loaded, and
moving right from the item pane only once an item is selected — otherwise
focus lands on an empty pane with no way to act.

Selection changes cascade: moving the vault cursor clears the item pane and
loads the new vault's items; moving the item cursor clears the detail pane and
loads the item. Loads are debounced by 150ms so holding `j` down does not
launch one `op` process per keystroke.

## Key routing

`Update` checks in this order, and stops at the first match:

1. An overlay is open — the form or confirmation gets the key, nothing else does.
2. The list is filtering — `list.Model` gets the key, so `/`-search accepts `j` and `h` as text.
3. A global binding — `Quit`, `Help`.
4. A pane binding — movement, CRUD, reveal, copy.

Without step 2, typing a vault name into the filter would move the cursor.

## Bindings

| Key | Action |
|---|---|
| `↑` `k` / `↓` `j` | move within pane |
| `←` `h` / `→` `l` | move focus between panes |
| `a` | create — vault, item, or field, per focused pane |
| `e` | edit selection |
| `d` | delete selection, after confirmation |
| `r` | reveal the concealed fields of the selected item |
| `y` | copy the selected item's password to the clipboard |
| `/` | filter the focused list |
| `?` | full help |
| `esc` | close overlay, or clear filter |
| `q` `ctrl+c` | quit |

`a`, `e`, and `d` are context-sensitive: they act on whatever the focused pane
holds. One key per verb, three meanings, resolved by focus — the alternative
is nine bindings for the same nine actions.
