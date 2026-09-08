# 2026-09-07 — Fixes from the first real run

The program was run against a real 1Password account for the first time. Four
findings, two of them bugs, fixed by two agents in parallel and integrated
here.

## Initial selection loaded nothing

On launch the first vault was highlighted but its items never loaded; the same
on entering a vault, where the first item's fields never loaded. Both needed
the cursor moved off and back on.

One root cause, in code written during
[wave 1 integration](03-00-00-wave-1-integration.md): `moveCursor` cascaded
only when the highlighted id *changed*. A list arriving with row 0 already
highlighted has no previous value to differ from, so nothing was ever
scheduled. The cascade was driven by the event, not by the state.

Loading is now driven by what is highlighted: a vault list or item list
arriving schedules a load for its highlighted row, batched behind `SetItems`.
`moveCursor` is untouched, so the debounce and the do-not-reload-when-the-
cursor-cannot-move behavior are unchanged.

Confirmed red before green twice — once by the agent, once here by reverting
the three source files to `origin/main` and watching the new tests fail with
"scheduled 0 loads, want 1". A reload-loop guard test came with it.

## The detail pane is now the item's edit form

The pane rendered a read-only `lipgloss/table` and editing happened in a
half-screen overlay. It is now a `huh` form over the item, blurred until `e`
focuses it, `esc` to leave discarding edits, `huh`'s own navigation inside.
`→`/`l` no longer enters the pane — `e` is the only way in, and the footer says
so. See [ADR 08](../adr/08-00-00-inline-item-editing.md).

Two consequences the agent surfaced rather than buried:

- **`r` moved to the item pane.** A focused form takes every printable key, so
  reveal could not stay on the detail pane.
- **`Quit` had to split.** It bound `q` and `ctrl+c` together and ran before
  pane handlers, so typing `q` into a form field would have quit the program.
  `Quit` is `q`; a new `Interrupt` is `ctrl+c`.

Masking survived the rewrite, which was the constraint that mattered:
concealed fields use huh's password echo. Verified here by deleting the echo
mode and watching 5 tests go red across two packages, then restoring.

## Empty-state padding

`select an item` rendered flush against the left border. It is now inset to
match the form's own field inset, with a test that measures the indent from
the border rather than asserting a string.

## Not bugs

- **No "Personal" vault.** `op vault list` itself returns five vaults and none
  is named Personal, Private, Employee, or Shared — all five are custom-named.
  `internal/op/vault.go` has no filter, skip, or exclude; the pane shows
  exactly what `op` returns.
- **`notesPlain`** is 1Password's own field label, passed through unchanged.

## Cleanup done at integration

Both agents reported dead code in files they did not own, and both correctly
left it alone. Removed here: `openEditItem` and the `keys.Edit` case in
`handleItemKey` (unreachable — `handleItemPaneDetailKey` claims Edit first),
and `theme.Mask`, which huh's password echo replaced. The two selection test
helpers collapsed to one implementation.

## Known regression, being fixed separately

The inline form edits field **values** and the item title. It cannot rename,
retype, add, or remove a field on an existing item — those remain only in the
create overlay. The original brief asked for CRUD on "names/fields/etc in an
entry", so this is a capability loss, not a scope decision, and a follow-up is
in flight. ADR 08 records what it would take.
