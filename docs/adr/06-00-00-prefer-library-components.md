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
| Detail pane scrolling | `bubbles/viewport` |
| Field alignment | `lipgloss/table` — it computes column widths |
| Loading indication | `list.StartSpinner` |
| Footer and `?` overlay | `bubbles/help`, generated from the keymap |
| Key bindings | `bubbles/key` |
| Create and edit forms | `huh` |
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

**`Detail`** — the field pane composes `viewport` and `lipgloss/table` rather
than using `bubbles/table`. `bubbles/table` is an interactive, row-selectable
grid; the detail pane is a read-and-scroll document with per-section grouping
and per-field masking. Using it would mean fighting its selection model for a
behavior that is not wanted.

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
