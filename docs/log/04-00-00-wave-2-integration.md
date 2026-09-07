# 2026-09-07 — Wave 2 built and integrated

Vault CRUD and item/field CRUD, built concurrently, then integrated. This
completes the CRUD scope: every `op` call the design named is implemented.

## Two agents, not three

[ADR 07](../adr/07-00-00-parallel-slice-construction.md) planned wave 2 as
three agents — vault, item, and field CRUD. It was two.

Field editing lives *inside* the item form, so a separate field agent would
have fought the item agent over one file. The partition rule that made wave 1
work — one owner per file — is what ruled the third agent out. ADR 07 already
says coordination is not worth it for small slices; this was that case.

## The contract had a bug that made every form unsubmittable

The vault agent found that **no `huh` form in the program could ever be
submitted**, and correctly reported it instead of editing the frozen file.

`Update` consulted `m.overlay` only inside `handleKey`, so key presses reached
an open form but nothing else did. `huh` advances on its own messages —
a field returns `NextField`, the runtime delivers `nextFieldMsg`, then
`nextGroupMsg`, and only then does `Form.State` become `StateCompleted`. Those
messages fell through the switch into `forwardToFocused` and were handed to
the focused *pane*.

Reproduced here with a test that drives a form through `Model.Update` rather
than reaching into `updateOverlay`, confirmed red, fixed, confirmed green, then
reverted the fix to confirm the test was catching the bug and not the harness.
The fix is four lines: an open overlay takes every message the explicit cases
did not claim.

This was a slice 0 defect, which is worth naming plainly: the contract froze a
seam that had never been exercised end to end. Slice 0's own bar — "it runs" —
was met by a program with no forms in it yet.

## The one merge collision was predicted

Both agents wrote a `drain` test helper in the `form` package. The vault agent
called it in its report before the merge happened.

Resolved by keeping the stronger version (it recurses on the command a form's
own update returns, and carries a budget so a form that never settles fails
rather than hangs) and deleting the weaker one. Its magic `16` became a named
`settleBudget`.

Nothing else conflicted. Both agents kept every identifier prefixed and neither
created a shared file in the package they co-owned.

## A test the merge invalidated

`TestCrudKeysDoNothingOnTheVaultPane` asserted that `a`, `e`, and `d` open no
overlay while the vault pane is focused. That was true when the item agent
wrote it — vault CRUD did not exist yet — and false the moment both branches
met.

The test's intent was right and its assertion was stale. It now asserts that
whatever opens on the vault pane is not an *item* form, which is the property
it was always trying to protect. Worth noting as a hazard of parallel work:
a green test on a branch can encode the absence of a feature another branch is
adding.

## Polish done here

- **A deleted record no longer lingers on screen.** `handleWriteSucceeded`
  routed only to the pane named for reload, so after a delete the detail pane
  kept showing the deleted item — including a concealed field the user had
  revealed. `writeSucceededMsg` gained `Removed`, and a delete now clears the
  panes below. Tested by revealing a secret, deleting, and asserting the value
  is gone from the rendered output.
- **`clearDownstream` became `clearBelow(pane)`.** The selection cascade and
  the delete path want the same operation from different triggers; two copies
  would have been the duplication rule 3 forbids.
- **`errNotImplemented` removed.** Every signature slice 0 froze is now
  implemented, so the sentinel had no referents left. Lint caught it.

## Design corrections from the agents

- **`huh` has no `EchoModeFunc`** — echo mode is fixed when an input is built,
  so a single value input cannot switch to password echo when the field type
  changes. The item agent renders plain and password inputs and hides the one
  that does not apply, so a concealed value is never echoed.
- **A repeating field editor is not a `huh` primitive.** Groups are static.
  Add/edit/delete of fields is one rule: the form carries the existing fields
  plus one blank row; filling the blank adds, editing in place edits, clearing
  a label removes.
- **Editing needs the loaded item.** `op item list` returns no field values, so
  editing off the item pane's selection alone would erase every field. Edit
  opens only when the detail pane holds the same item.
- **`op vault edit` writes nothing to stdout**, so `EditVault` composes its
  return from its arguments. The frozen signature implied a decode with no
  source.
- Components 01 and 05 were amended by the item agent in the same commit.

## Verified

`mise run build`, `mise run lint` (0 issues), and `go test ./... -count=1` are
clean on the merged tree.

The security property has a real test: `TestItemWritesNeverPutAValueInArgv`
records the argv a fake `op` receives across create, edit, and delete, and
asserts no field value appears in any argument and no non-flag argument
contains `=`. Values reach `op` only through stdin. Nothing is written to a
temp file. See [ADR 04](../adr/04-00-00-secrets-never-in-argv.md).

## Not yet done

The program has never been run against a real 1Password account. Every test
injects `Client.Path` at a fake `op`, and every fixture is synthetic — which is
the right default, but it means the first real `op vault list` is still
unproven. That is the next thing worth doing, and it needs a human at the
keyboard for the sign-in.

Remaining from the design: clipboard copy (`y`), and the overlay floating over
the panes rather than replacing them.
