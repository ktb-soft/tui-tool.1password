# 2026-09-07 — Slice 0, the spine

Slice 0 of [ADR 07](../adr/07-00-00-parallel-slice-construction.md). The
contract every later agent codes against, and none of them changes.

## What exists now

`go.mod` pins `bubbletea/v2 v2.0.8`, `bubbles/v2 v2.1.1`, `lipgloss/v2 v2.0.5`
and `huh/v2 v2.0.3`. `internal/op` has every type and every exported
signature, with `Client.run`, `OpError`, and the `list.Item` methods real and
the rest returning `errNotImplemented`. `internal/ui/theme` and
`internal/ui/keymap` are complete. `internal/app` has the messages, the root
model, the dispatch switch, and the six per-vertical handlers as no-op stubs.
`testdata/` holds hand-written fixtures.

`mise run build`, `mise run lint` and `go test ./...` are clean, and
`./bin/optui` draws three empty bordered panes and quits on `q`.

## Fixtures are synthetic

`testdata/` was hand-written against the shapes `op --format=json` documents,
not recorded from a live vault. Every value is fake — `example.com`,
`hunter2`, repeated-character IDs. Recording real output would have put real
credentials in git history, which is the thing
[ADR 04](../adr/04-00-00-secrets-never-in-argv.md) exists to prevent.

This is a deliberate departure from ADR 07, which asked for "real
`op --format=json` output ... recorded once and scrubbed". Scrubbing is a
manual step that fails open; synthesis fails closed.

## Where the design was wrong

Four things in the design did not survive contact, and are amended in place.

**A Go field and method cannot share a name.**
[Component 02](../components/02-00-00-domain-model.md) gave `Item` a `Title`
field and a `Title()` method, and `Vault` a `Description` field and a
`Description()` method. Neither compiles. The fields are now `Item.Name`
(`json:"title"`) and `Vault.Note` (`json:"description"`), which keeps both
types on `list.DefaultItem` and lets both panes share one delegate.

**`huh.Form` is not a `tea.Model` in v2.** Its `Update` returns
`(huh.Model, tea.Cmd)`. [Component 05](../components/05-00-00-crud-forms.md)
said it drops in with no adapter; it does, but the field has to be typed
`*huh.Form`, not `tea.Model`.

**`lipgloss.Style.Width` is the outer width.**
[Component 03](../components/03-00-00-pane-view.md) said to subtract the
border width before sizing the pane, which makes the three columns two cells
short each. Only `list.SetSize` subtracts it now.

**`help.Model` does not truncate to its width.** With ten bindings the footer
rendered 107 cells inside an 80-cell terminal, and `JoinVertical` padded every
pane row out to match. `ShortHelp` is now the seven bindings the mock in
`docs/index.md` shows, and `theme.Footer` caps the row with
`lipgloss.MaxWidth` as a backstop.

**`huh` does not abort on `esc`.** Its default keymap binds only `ctrl+c`
(`huh/keymap.go:109`), but
[component 04](../components/04-00-00-navigation-keymap.md) documents `esc` as
"close overlay". `update.go` matches `KeyMap.Cancel` before handing the key to
the form, so the documented behavior is the contract's, not `huh`'s.

## Contract additions ADR 07 did not name

`internal/ui/pane` and `internal/ui/detail` are created here, not in wave 1.
`app.go` holds a `pane.Pane` and a `detail.Detail`, so without them every
wave-1 agent adds the same field to `app.go` and collides in the one file the
partition exists to protect. The files pass to their owning verticals intact.

`Model.overlaySubmit` is new. The overlay is opened by a vertical but closed by
`update.go`, and without a submit hook `update.go` would need a branch per
form — which is exactly the edit ADR 07 forbids. `openOverlay` goes with it,
so every vertical sizes and dismisses its forms the same way instead of six
agents inventing six conventions.

`selectionChangedMsg` and `debounceSelection` are new, carrying the 150ms
debounce [component 04](../components/04-00-00-navigation-keymap.md)
specifies. `writeSucceededMsg` gained a `Status` field so the status-bar text
comes from the write that produced it.

## Left for later

The overlay is centered with `lipgloss.Place`, which is what
[component 05](../components/05-00-00-crud-forms.md) specifies and means the
form covers the panes rather than floating over them. `lipgloss/v2` does ship
`Canvas` and `Layer`, but composing the panes and a positioned form with them
did not work on the first two obvious spellings, and freezing a seam that does
not work is worse than freezing the simple one. If wave 2 wants a true float,
that is a change inside `render`, which no vertical owns.

## Next

Wave 1: the vault, item, and detail panes, in parallel.
