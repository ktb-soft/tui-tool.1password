# Inline item editing in the detail pane

## Status

Accepted. Amends [ADR 03](03-00-00-column-focus-navigation.md), which said
`→`/`l` advances into the detail pane, and narrows the "Create and edit forms"
row of [ADR 06](06-00-00-prefer-library-components.md) from "overlay" to
"overlay or detail pane".

## Context

The detail pane rendered the selected item through `lipgloss/table` inside a
`viewport`: a read-only document. Editing an item meant `e`, which replaced the
whole screen with a `huh` overlay paging one field at a time. Two consequences:

- **The third pane could be focused but did nothing.** `→` from the item pane
  moved focus onto a viewport that had no cursor and no verbs. The footer
  advertised a move that bought the user nothing — the phantom-feature problem
  [the wave 2 log](../log/04-00-00-wave-2-integration.md) already had to
  resolve once, for the clipboard binding.
- **Editing lost the context it was editing.** The overlay covered the pane
  showing the values being changed, and paged the fields so only one was
  visible at a time. The pane and the form drew the same data twice, in two
  layouts, and never at once.

## Decision

**The detail pane is the item's edit form.** It holds a `huh.Form` built by
`form.NewInlineItem`: the item's title over one input per field, all in a
single `huh.Group` so the whole item is visible rather than paged.

**`e` moves focus into it; `→`/`l` no longer does.** `→` advances only from the
vault pane to the item pane. There is one way into the detail pane and it is
the key that already meant "edit this".

**The pane is blurred by default.** It renders the item without taking keys.
Reading an item is the common case and must not require entering an editor.

**`esc` leaves without saving.** `Focus` and `Blur` both rebuild the form from
the stored item, so a blur discards by construction rather than by a discard
path that has to be remembered.

**Navigation inside the pane is `huh`'s.** `tab`/`enter`/`shift+tab` are the
form's own bindings; nothing here reimplements them. Only `esc` and `ctrl+c`
are intercepted before the form sees them.

**Concealed fields use `huh.EchoModePassword`.** `huh` fixes an input's echo
mode at construction, so `r` toggles a flag and rebuilds the form. Reveal is
pressed on the item pane, because a focused form takes every printable key.

Item **create**, item **delete**, and all vault CRUD keep the overlay.

## Consequences

`Quit` had to split. It bound `q` and `ctrl+c` together, and the global handler
ran before the pane handlers — so with a form in a pane, typing `q` into a
field would have quit the program. `Quit` is now `q`, `Interrupt` is `ctrl+c`,
and the detail pane honours only the latter.

`Detail` no longer measures or groups anything: `groupFields` and the
`lipgloss/table` rendering are gone, and with them `internal/ui/detail/fields.go`.
Section headings survive as a `section · label` prefix on the field's title.

**Structural field editing moved out of edit.** The inline form has no label
input, no type select, and no blank row, so an existing item's fields can have
their values and the item's title changed, but cannot be renamed, retyped,
added, or removed. Those are still available where the fields are first
defined — the create overlay. The alternative was three rows per field in the
narrowest column, which would have had to page, which is what this ADR removes.
If per-field structure editing is wanted after create, it comes back as its own
decision with its own surface, not by widening this form until it pages again.

The pane shows the values it submitted while the write is in flight rather than
the ones it was seeded with. A rejected write therefore leaves the pane
optimistic until the next load; the `op` error is still shown verbatim in the
status bar.

Rejected: keeping `→` as a second way in, which would have left the same
phantom binding pointing at a pane that now starts an edit; and rendering the
form read-only when blurred and swapping in an editable copy on focus, which is
two renderings of one thing — the problem this ADR exists to remove.
