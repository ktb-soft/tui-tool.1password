# Column focus model

## Status

Accepted.

## Context

Three panes, two key families for each direction (arrows and vim), and a
right-hand pane that is a scrolling document rather than a list. Focus has to
be unambiguous at every moment, or `d` deletes the wrong thing.

## Decision

Focus is one integer on the app model, `0..2`, clamped rather than wrapped.
`h`/`←` decrement, `l`/`→` increment.

No wrap: `h` in the leftmost pane does nothing. Wrapping would move the cursor
to the far side of the screen from where the user is looking, and there is no
"jump to the end" gesture that a wrap would be serving.

Advancing right is gated on the destination having content — the item pane
needs a loaded vault, the detail pane a loaded item.

CRUD verbs are context-sensitive rather than per-pane: `a`, `e`, `d`, and `/`
act on the focused pane's entity.

## Consequences

Nine CRUD actions are covered by three keys, and the footer shows three
bindings instead of nine. The cost is that a key means different things
depending on focus, so the focused border must be unmistakable: it is drawn in
the accent color, and both focused and blurred borders use the same character
set so nothing shifts by a cell when focus moves.

Rejected: tab-cycling focus, which discards the spatial mapping the arrow keys
already provide; and per-pane verbs such as `av`/`ai`/`af`, which triple the
bindings without removing any ambiguity the visible focus border does not
already remove.
