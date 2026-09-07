# tui-tool.1password

A TUI Tool for streamlined interaction with 1Password in the terminal.

Three panes: vaults, the items in the selected vault, and the fields of the
selected item. Arrow keys or `hjkl` to move. `a`, `e`, `d` to create, edit, and
delete whatever the focused pane holds.

## Requirements

The [1Password CLI](https://developer.1password.com/docs/cli/) (`op`), signed
in. This tool runs `op` and inherits your session — it stores no credentials of
its own.

## Setup

```sh
mise install
```

That installs the pinned toolchain, installs the git hooks, and restores the
repo's agent context from `.config/kstore`. Nothing else is needed.

`mise run build`, `mise run test`, `mise run lint`.

## Documentation

[`docs/index.md`](docs/index.md) — how it works, the components, and the
decisions behind them.

## Status

Design complete, implementation not started. See
[the build order](docs/adr/05-00-00-vertical-slice-build-order.md).
