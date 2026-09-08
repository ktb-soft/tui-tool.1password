# CRUD forms

Every form in the program is built by `internal/ui/form`, which does not submit
any of them. There are two surfaces:

- an **overlay** over the three panes — vault create, vault edit, item create,
  item field structure, and every delete confirmation;
- the **detail pane** — item value editing. See
  [ADR 08](../adr/08-00-00-inline-item-editing.md).

## Overlay

The app model holds `overlay *huh.Form` (nil when none) and
`overlaySubmit func(Model) (tea.Model, tea.Cmd)`. While the overlay is
non-nil, `Update` routes every key to it and `View` places it centered over
the panes with `lipgloss.Place`.

In v2 a `huh.Form` does **not** satisfy `tea.Model`: its `Update` returns
`(huh.Model, tea.Cmd)`, not `(tea.Model, tea.Cmd)`. The field is therefore
typed `*huh.Form` rather than `tea.Model`, which also removes a type
assertion `update.go` would otherwise need.

`update.go` watches `overlay.State`. On `huh.StateCompleted` it clears the
overlay and calls `overlaySubmit`; on `huh.StateAborted` it clears the overlay
and does nothing else. The vertical that opened the form supplies
`overlaySubmit`, so `update.go` never learns which form is open and no
vertical has to edit it.

## Forms

**Vault** — one text input, bound to the name.

```go
huh.NewForm(huh.NewGroup(
	huh.NewInput().Title("Vault name").Value(&name).Validate(required),
))
```

Create starts empty; edit starts with the selected vault's name. Same
constructor, different seed value.

**Item (overlay)** — title, a category select, and one page per field. One
constructor serves two openings: `a` on the item pane opens it over an empty
item to create one, `f` opens it over the loaded item to change its field
structure. The field rules below are therefore written once.

```go
huh.NewForm(append([]*huh.Group{headerGroup}, fieldGroups...)...)
```

Category is selectable, because a create is the only time `op` will set it.

**Field** — label, a type select (`STRING` / `CONCEALED` / `URL` / `OTP`), and
a value, laid out as three groups per field: the label and type together, then
the value twice, once plain and once with `.EchoMode(huh.EchoModePassword)`.
`WithHideFunc` shows only the one matching the chosen type, because `huh` fixes
an input's echo mode at construction and offers no `EchoModeFunc`. A concealed
value therefore never appears on screen while typing, even when the type is
changed mid-form.

The form carries one blank row per field slot. Filling the blank row adds a
field; clearing a label removes that field, which the label's description
states. There is no separate add or delete action, and a field with an empty
value is kept.

**Item (detail pane, edit only)** — `form.NewInlineItem` builds one `huh.Group`
holding the item's title over one input per field, each titled with the field's
label and bound to its value. One group, so the whole item is visible at once
rather than paged, and `huh` navigates between the fields itself.

```go
huh.NewForm(huh.NewGroup(fields...))
```

A concealed field is constructed with `EchoMode(huh.EchoModePassword)` unless
the item has been revealed, so its value never renders. Because `huh` fixes an
input's echo mode at construction, revealing rebuilds the form.

The inline form edits **values and the title**, not field structure: it has no
label input, no type select, and no blank row. A per-field label-and-type
editor does not fit the third column without paging it — which is the thing the
redesign removed — so structure is edited in the overlay form above, reached
with `f`. Both openings of that form are the same constructor, so adding,
renaming, retyping, and removing a field mean the same thing whenever they are
done.

## Delete

Deletes use a `huh.NewConfirm` (`charm-huh.md:3978`) with the name of the
target in the prompt — `Delete vault "Shared"?` — defaulted to No. Vault
deletion additionally requires the vault to be typed back, because it takes
every item with it.

## Submission

A completed form produces a value, and the controller turns that value into a
`tea.Cmd`. The form never runs `op` itself, so the whole form package is pure
and testable by constructing it, feeding keys, and reading the bound
variables.

The detail pane submits the same way. `forwardToDetail` watches
`Detail.Completed`, reads `EditedItem`, and issues the same `saveItem` command
the overlay used — the write path is shared, only the surface differs. The pane
keeps the values it submitted while the write is in flight, so a saved edit
does not blink back to the value it was seeded with.

Item writes serialize the full `op.Item` to JSON and pipe it to
`op item create --vault <id> -` / `op item edit <id> -`. See
[the ADR](../adr/04-00-00-secrets-never-in-argv.md).
