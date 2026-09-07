# Charm v2 libraries

## Status

Accepted.

## Context

The Charm libraries have two live major versions. v1 is
`github.com/charmbracelet/bubbletea`; v2 is `charm.land/bubbletea/v2`. Most
published examples, and most recalled API knowledge, describe v1.

The APIs differ in ways that surface as confusing compile errors rather than
obvious ones:

- `View()` returns `tea.View`, not `string` (`charm-bubbletea.md:8183`)
- key presses arrive as `tea.KeyPressMsg`, not `tea.KeyMsg` (`charm-bubbletea.md:8161`)
- module paths are `charm.land/*`, not `github.com/charmbracelet/*` (`charm-bubbletea.md:10134`)

## Decision

Use v2 throughout — `bubbletea/v2 v2.0.8`, `bubbles/v2 v2.1.1`,
`lipgloss/v2 v2.0.5`, `huh/v2` — pinned to exact versions.

## Consequences

v1 snippets found online will not compile without translation. The upgrade
guide at `charm-bubbletea.md:13277` is the reference for the differences.

`bubbles/v2` and `huh/v2` both depend on `bubbletea/v2` and `lipgloss/v2`, so
mixing majors was never really available — a single v1 component would pull in
a second, incompatible event loop type.
