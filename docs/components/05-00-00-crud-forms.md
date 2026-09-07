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

**Item** — title, a category select, and a repeating field editor.

```go
huh.NewForm(
	huh.NewGroup(
		huh.NewInput().Title("Title").Value(&title),
		huh.NewSelect[string]().Title("Category").Options(categories...).Value(&category),
	),
	huh.NewGroup(fieldInputs...),
)
```

Category is fixed at create and read-only at edit — `op` will not change an
item's category.

**Field** — label, a type select (`STRING` / `CONCEALED` / `URL` / `OTP`), and
a value. Concealed fields use `.EchoMode(huh.EchoModePassword)` so the value
does not appear on screen while typing.

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
`op item create -` / `op item edit <id> -`. See
[the ADR](../adr/04-00-00-secrets-never-in-argv.md).
