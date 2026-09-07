# 2026-09-07 — Initial design

The repository held only a README. This entry records the design laid down
before any code exists.

## Scope agreed

Three horizontal panes — vaults, items, fields — with arrow and vim
navigation and full CRUD on vaults, items, and fields.

## What was verified, not assumed

- `op 2.36.0-beta.03` is installed at `/usr/local/bin/op`.
- `op vault` and `op item` both expose `create`, `get`, `edit`, `delete`, `list`.
- `op item edit` takes assignment statements of the form
  `[<section>.]<field>[[<fieldType>]]=<value>`, and its help warns that command
  arguments are visible to other local processes.
- `op item create` and `op item edit` both accept a JSON template piped on
  stdin with `-` as the first argument. This is what
  [ADR 04](../adr/04-00-00-secrets-never-in-argv.md) rests on.
- The Charm libraries in the local documentation mirror are v2 under
  `charm.land/*`, not v1 under `github.com/charmbracelet/*`
  (`charm-bubbletea.md:10134`). This changes `View()`, key messages, and every
  import path, and is why [ADR 01](../adr/01-00-00-charm-v2-libraries.md)
  exists.

## Open questions

- Multiple 1Password accounts. The design threads `Client.Account` through
  every call but has no account picker. If the tool is used against more than
  one account, that becomes a fourth pane or a startup select.
- Item categories other than Login. The field editor is designed against the
  generic field list, so other categories should work, but only Login has been
  reasoned through concretely.
- Archived items. `op item delete` can archive instead of deleting; the
  design currently deletes.

## Next

Slice 1 of [the build order](../adr/05-00-00-vertical-slice-build-order.md):
the vault pane.
