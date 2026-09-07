# 2026-09-07 — Design review

Review of [the initial design](01-00-00-initial-design.md) against a fifth
project rule that was not in scope when it was written: *don't reinvent the
wheel*. Two follow-up questions from the same review are recorded here.

## Rule 5 added

The project rules now live in `.claude/rules/`, and rule 5 says to use an
existing library, framework, or component where one provides what is needed
without sacrificing much.

## What the review changed

Three places in the design hand-rolled something already available:

- **Field alignment.** The detail pane described "an aligned two-column list",
  which means measuring strings and padding cells. Replaced with
  `lipgloss/v2/table`, which computes column widths itself
  (`charm-lipgloss.md:4064`).
- **Loading and empty states.** Each pane was to render its own centered
  `loading…` line from a theme constant. `list.Model` already has
  `StartSpinner` (`charm-bubbles.md:2185`) and `SetStatusBarItemName`
  (`charm-bubbles.md:1931`). The panes configure those instead.
- **Status line.** The app model carried a `status string` that the footer
  rendered. Replaced with `list.NewStatusMessage` (`charm-bubbles.md:2202`),
  which expires on its own, removing both the field and the footer branch.

The two components that remain hand-rolled — `Pane` and `Detail` — now have
their reasons recorded in
[ADR 06](../adr/06-00-00-prefer-library-components.md).

## Questions answered

- **Charm v2 confirmed.** The documentation mirror is v2 throughout and
  matches what the design targets. Nothing in it is stale.
- **Context-sensitive actions confirmed.** `a`/`e`/`d` act on the focused
  pane's entity, as designed in
  [ADR 03](../adr/03-00-00-column-focus-navigation.md).
- **Piping secrets over stdin confirmed** as the write path
  ([ADR 04](../adr/04-00-00-secrets-never-in-argv.md)).
- **Parallel construction** — asked whether the verticals could be built
  concurrently by several agents. They can, given a frozen contract and a
  partition by file rather than by package; this became
  [ADR 07](../adr/07-00-00-parallel-slice-construction.md), which adds a
  serial slice 0 ahead of the seven slices and groups the rest into two
  parallel waves.

## Next

Slice 0 of [ADR 07](../adr/07-00-00-parallel-slice-construction.md): the
spine — types, signatures, constants, messages, dispatch, and recorded
`testdata/`. Serial, one agent, and nothing else starts until it compiles and
runs.
