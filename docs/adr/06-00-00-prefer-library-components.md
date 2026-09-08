# Prefer library components

## Status

Accepted.

## Context

A TUI is mostly solved problems: scrolling lists, filtering, pagination,
column alignment, text input, spinners, form validation, help generation.
`bubbles`, `lipgloss`, and `huh` are already dependencies and already ship all
of it.

Hand-rolling any of them produces code that has to be maintained, tested, and
read, in exchange for nothing.

## Decision

Use the library component wherever one covers the need. Specifically:

| Need | Component |
|---|---|
| Vault and item panes | `bubbles/list` |
| Filtering, pagination, status messages | `list` built-ins, not custom code |
| Detail pane | `huh` — the pane is the item's edit form |
| Loading indication | `list.StartSpinner` |
| Footer and `?` overlay | `bubbles/help`, generated from the keymap |
| Key bindings | `bubbles/key` |
| Create and edit forms | `huh`, in an overlay or in the detail pane |
| Delete confirmation | `huh.NewConfirm` |
| Input validation | `huh` validators |
| Layout | `lipgloss.JoinHorizontal`, `JoinVertical`, `Place` |

No code in this project measures a string, pads a cell, tracks a scroll
offset, or implements fuzzy matching.

## Where this design does not

Two components are written here rather than taken:

**`Pane`** — a thin wrapper holding a `list.Model`, a title, and a focus flag.
It is not a reimplementation of a list; it is the border-and-focus decoration
around one, roughly thirty lines, and it exists so the two list panes are one
type instead of two. Taking a dependency for this would be the more complex
option.

**`Detail`** — the field pane is a thin holder for a `huh.Form`, roughly the
same shape as `Pane`: it owns the border, the focus flag, the reveal flag, and
the rebuild that discards edits. The form, its layout, its scrolling, and its
navigation are all `huh`'s. Nothing here is a reimplementation of one.

This entry used to describe a `viewport` and `lipgloss/table` composition; see
[ADR 08](08-00-00-inline-item-editing.md) for why the pane became a form.

Everything else comes from a library. When a future component is hand-rolled,
it gets an entry here saying why — an unexplained hand-rolled component is
complexity that was never justified.

## Consequences

Much less code, and the parts most likely to have edge-case bugs — wrapping,
truncation, resize, unicode width — are someone else's tested code.

The cost is inheriting the libraries' opinions. `list` decides what filtering
feels like and `huh` decides what a form looks like. Both are configurable
enough that this has not yet forced a compromise; if one does, the escape is
to replace that single component, which the one-way dependency structure
already allows.
