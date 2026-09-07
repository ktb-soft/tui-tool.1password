# optui

A three-pane terminal UI over the 1Password CLI (`op`). Vaults on the left,
items in the middle, the selected item's fields on the right. Full CRUD on all
three.

```
┌ Vaults ───────┬ Items ──────────────┬ Personal Website ─────────────┐
│ > Personal    │ > GitHub            │ category  LOGIN               │
│   Shared      │   Personal Website  │ username  brian@example.com   │
│   Archive     │   Router Admin      │ password  ••••••••••  (r)     │
│               │   AWS Root          │ url       https://example.com │
│               │                     │                               │
│               │                     │ Notes                         │
│               │                     │ renewal 2027-01               │
└───────────────┴─────────────────────┴───────────────────────────────┘
 h/l ←→ pane   j/k ↑↓ move   a add   e edit   d delete   r reveal   ? help
```

## Architecture

Model / View / Controller, one Go package per concern.

```
cmd/optui/main.go          program entrypoint

internal/op/               MODEL — data + the only place that shells out to `op`
  client.go                Client: exec wrapper, JSON in/out
  vault.go                 Vault, ListVaults, CreateVault, EditVault, DeleteVault
  item.go                  Item, Field, ListItems, GetItem, CreateItem, EditItem, DeleteItem

internal/ui/               VIEW — rendering only, no `op` calls
  theme/theme.go           every color, border, and dimension constant
  keymap/keymap.go         every key binding, and the help text derived from it
  pane/pane.go             Pane: bordered, focusable list container
  detail/detail.go         Detail: the field pane
  form/form.go             huh forms for create/edit, one per entity

internal/app/              CONTROLLER — wiring
  app.go                   Model: focus, panes, overlay, size
  update.go                message routing
  commands.go              tea.Cmd wrappers around internal/op
  messages.go              vaultsLoadedMsg, itemsLoadedMsg, opFailedMsg, ...
```

Dependency direction is one-way: `app` imports `ui` and `op`; `ui` imports
neither `app` nor `op`. The panes render values handed to them, so a pane can be
replaced without touching data code, and `op` can be swapped for the Go SDK
without touching rendering.

## Libraries

Charm v2 (`charm.land/*`), pinned exactly.

| Library | Use |
|---|---|
| `charm.land/bubbletea/v2` | event loop, `Model`/`Init`/`Update`/`View` |
| `charm.land/bubbles/v2/list` | vault and item panes |
| `charm.land/bubbles/v2/viewport` | detail pane scrolling |
| `charm.land/bubbles/v2/key` | binding definitions |
| `charm.land/bubbles/v2/help` | footer, generated from the keymap |
| `charm.land/bubbles/v2/spinner` | in-flight `op` calls, driven by `list.StartSpinner` |
| `charm.land/lipgloss/v2` | borders, colors, `JoinHorizontal` |
| `charm.land/lipgloss/v2/table` | field rendering in the detail pane |
| `charm.land/huh/v2` | create/edit forms, confirmations, validation |

Nothing here is hand-rolled that one of these already provides — filtering,
pagination, status messages, spinners, help generation, and column alignment
are all taken from the libraries. See
[ADR 06](adr/06-00-00-prefer-library-components.md) for the two places the
design deliberately does not.

v2 signatures differ from v1: `Update(tea.Msg) (tea.Model, tea.Cmd)`, `View()`
returns `tea.View` (`charm-bubbletea.md:8158`, `:8183`), and key presses arrive
as `tea.KeyPressMsg`, not `tea.KeyMsg` (`charm-bubbletea.md:8161`).

## Components

- [op client](components/01-00-00-op-client.md) — the `op` subprocess adapter
- [Domain model](components/02-00-00-domain-model.md) — Vault, Item, Field
- [Pane view](components/03-00-00-pane-view.md) — three-column layout and focus
- [Navigation and keymap](components/04-00-00-navigation-keymap.md)
- [CRUD forms](components/05-00-00-crud-forms.md) — huh overlays
- [App controller](components/06-00-00-app-controller.md) — root model and messages

## Decisions

- [Charm v2](adr/01-00-00-charm-v2-libraries.md)
- [`op` CLI subprocess, not the Go SDK](adr/02-00-00-op-cli-not-sdk.md)
- [Column focus model](adr/03-00-00-column-focus-navigation.md)
- [Secrets never in argv](adr/04-00-00-secrets-never-in-argv.md)
- [Vertical-slice build order](adr/05-00-00-vertical-slice-build-order.md)
- [Prefer library components](adr/06-00-00-prefer-library-components.md)
- [Parallel slice construction](adr/07-00-00-parallel-slice-construction.md)

## Build order

One vertical at a time — model, view, and controller for a single pane, built
and working before the next starts. See
[the ADR](adr/05-00-00-vertical-slice-build-order.md) for why, and
[the log entry](log/01-00-00-initial-design.md) for the slice list.
