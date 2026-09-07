# CRUD forms

Creates and edits happen in a `huh` form rendered over the three panes.
`internal/ui/form` builds the forms; it does not submit them.

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

**Item** — title, a category select, and one page per field.

```go
huh.NewForm(append([]*huh.Group{headerGroup}, fieldGroups...)...)
```

Category is selectable at create and rendered as a read-only `huh.NewNote` at
edit — `op` will not change an item's category.

**Field** — label, a type select (`STRING` / `CONCEALED` / `URL` / `OTP`), and
a value, laid out as three groups per field: the label and type together, then
the value twice, once plain and once with `.EchoMode(huh.EchoModePassword)`.
`WithHideFunc` shows only the one matching the chosen type, because `huh` fixes
an input's echo mode at construction and offers no `EchoModeFunc`. A concealed
value therefore never appears on screen while typing, even when the type is
changed mid-form.

The form carries the item's existing fields plus one blank row. Filling the
blank row adds a field; clearing a label removes that field, which the label's
description states. There is no separate add or delete action, and a field with
an empty value is kept.

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

Item writes serialize the full `op.Item` to JSON and pipe it to
`op item create --vault <id> -` / `op item edit <id> -`. See
[the ADR](../adr/04-00-00-secrets-never-in-argv.md).
