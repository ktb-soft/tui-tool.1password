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
	Reveal         key.Binding
	Filter, Help   key.Binding
	Confirm, Cancel key.Binding
	Quit, Interrupt key.Binding
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

`→`/`l` advances only from the vault pane to the item pane, and only once
items have loaded. **The detail pane is entered with `e`, not with an arrow** —
it is a form, and moving into it means starting an edit. `esc` leaves it and
discards whatever was typed. See
[ADR 08](../adr/08-00-00-inline-item-editing.md).

While the detail pane holds focus, `huh` navigates itself: `tab`/`enter`
advance a field, `shift+tab` goes back, and completing the last field submits.
Only `esc` and `ctrl+c` are intercepted, so the global `q` and `?` do not steal
a keystroke from an input. That is why `Quit` binds `q` alone and `Interrupt`
binds `ctrl+c` — one binding for both would make the program unquittable from
inside a form, or the letter `q` untypeable.

Selection changes cascade: moving the vault cursor clears the item pane and
loads the new vault's items; moving the item cursor clears the detail pane and
loads the item. Loads are debounced by 150ms so holding `j` down does not
launch one `op` process per keystroke.

## Key routing

`Update` checks in this order, and stops at the first match:

1. An overlay is open — the form or confirmation gets the key, nothing else does.
2. The list is filtering — `list.Model` gets the key, so `/`-search accepts `j` and `h` as text.
3. The detail pane has focus — `esc` leaves, `ctrl+c` quits, everything else is `huh`'s.
4. A global binding — `Quit`, `Interrupt`, `Help`.
5. A pane binding — movement, CRUD, reveal.

Without step 2, typing a vault name into the filter would move the cursor.

## Bindings

| Key | Action |
|---|---|
| `↑` `k` / `↓` `j` | move within pane |
| `←` `h` / `→` `l` | move focus between the vault and item panes |
| `a` | create — a vault or an item, per focused pane |
| `e` | edit — a vault in an overlay, an item in the detail pane |
| `d` | delete selection, after confirmation |
| `r` | reveal the concealed fields of the loaded item |
| `/` | filter the focused list |
| `?` | full help |
| `esc` | leave the detail form without saving, close an overlay, or clear a filter |
| `q` | quit |
| `ctrl+c` | quit, including from inside a form |

`a`, `e`, and `d` are context-sensitive: they act on whatever the focused pane
holds. One key per verb, two meanings, resolved by focus — the alternative is
a binding per pane per verb.

`e` is the only asymmetry, and it is deliberate: on the vault pane it opens an
overlay, on the item pane it moves focus into the detail pane's form. Both are
"edit the selection"; only the surface differs.
