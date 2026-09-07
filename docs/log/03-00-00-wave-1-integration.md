# 2026-09-07 — Wave 1 built and integrated

The three read-path verticals — vaults, items, detail — were built
concurrently by three agents against the slice 0 contract, per
[ADR 07](../adr/07-00-00-parallel-slice-construction.md). This entry records
what the experiment actually produced.

## The partition worked

Three branches merged into one tree with **zero conflicts**. No agent touched
a file another agent owned, and none needed a frozen file changed.

The two additions that made it work were not in ADR 07 and belong there for
the next wave:

- **`theme` was a three-way collision waiting to happen.** Every vertical
  wants new constants in one package. Each agent was given its own new file in
  that package (`theme/vaults.go`, `theme/items.go`, `theme/detail.go`) rather
  than a shared file to edit. Additive, same package, no conflict. As it
  happened none of the three needed one.
- **`ui/pane` and `ui/detail` already existed** from slice 0, which ADR 07 did
  not anticipate. Agents were told they were extending, not creating.

## The cascade, written serially afterwards

`internal/app/selection.go` is the one genuinely cross-vertical piece, and it
was written here rather than by any agent — which is what ADR 07 called for.

Movement keys are intercepted in `handlePaneKey`, forwarded to the focused
list, and the highlighted id compared before and after. Only a real change
cascades: it clears the panes downstream of the moved cursor and schedules a
debounced load. Pressing `k` at the top of a list changes nothing and must not
issue an `op` call.

Clearing downstream matters for more than tidiness — without it a previous
item's fields sit beside a newly selected vault while the load is in flight,
which is a correctness problem, not a cosmetic one.

Both halves of the receiving side already existed: the item agent handled
`selectionChangedMsg` in its own file (correctly — it owned the file and had
been told to handle every message routed there), and only the vault half and
the emit side were left.

## Three dead constants removed

`theme.NoVaults`, `theme.NoItems`, and `theme.SecretVal` had no consumers.

The first two duplicated `bubbles/list`'s own empty state, which `pane.New`
already configures through `SetStatusBarItemName` — rule 5 says the library
owns that string, so a second copy is exactly the duplication rule 3 forbids.
`SecretVal` could not be used without the detail package tracking row indices
to style one cell, which is the measuring and positioning rule 5 forbids.

Two agents found these independently and both left them alone and reported,
because they sat in a frozen file. That is the behavior the freeze was meant
to produce.

## Design corrections from the agents

- **Reveal is not per-key-press per-field.** Documented in
  [component 03](../components/03-00-00-pane-view.md): state is per field, `r`
  toggles the item. The detail agent also fixed an inherited bug where
  toggling flipped each field independently, leaving mixed states.
- **`op` list rows carry more than the model keeps** —
  `additional_information`, `urls`, `password_details` are dropped. Fine for
  the read path.
- **`version` and `last_edited_by` are unmodeled.** This is the open question
  for wave 2: if `op item edit` wants a `version` round-tripped for optimistic
  concurrency, item CRUD needs it before it starts.

## Next

Wave 2, three agents: vault CRUD, item CRUD, field CRUD. Resolve the `version`
question first — it belongs to the contract, not to a vertical.
