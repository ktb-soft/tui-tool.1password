# 2026-09-07 — The first row of a list was never loaded

On launch the vault pane highlighted its first vault and showed nothing beside
it. The items appeared only after moving the cursor off that vault and back
onto it. The item pane had the same fault: the first item never showed its
fields.

## Root cause

`moveCursor` cascaded on the highlight *changing*, comparing the selected id
before and after forwarding a movement key. A list that arrives already
highlighting row 0 never goes through `moveCursor`, so nothing was ever
scheduled for it. The cascade was driven by the event, not by the state.

## Fix

`scheduleLoad(pane)` in `selection.go` reads whatever the named pane currently
highlights and returns the same debounced `selectionChangedMsg` `moveCursor`
does. `vaultsLoadedMsg` and `itemsLoadedMsg` now batch it behind the
`SetItems` that populated the pane, so the cascade is driven by what is
highlighted rather than by the highlight changing.

`moveCursor` is untouched: the debounce still holds, and a movement key that
cannot move still issues no `op` call.

## Why this does not loop

Each pane schedules a load only for the pane it just filled, and an empty pane
highlights nothing and schedules nothing. The chain is finite: vaults → items
→ one item's fields, and the fields land in the detail pane, which is not a
list and populates nothing further. `TestLoadingAnItemDoesNotRepopulateTheItemPane`
holds that last link.

The tests were confirmed red against the old cascade before the fix and green
after.
