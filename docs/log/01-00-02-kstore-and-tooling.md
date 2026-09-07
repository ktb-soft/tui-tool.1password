# 2026-09-07 — kstore and tooling wired up

`CLAUDE.md` and `.claude/` are globally gitignored, so the project rules
written during [the design review](01-00-01-design-review.md) existed on one
disk and nowhere else. This wires them into the repo the way kstore intends.

## What was added

**`mise.toml`** — pins `go 1.27.1`, `golangci-lint 2.13.2`, `lefthook 2.1.12`,
and `age 1.3.2` (kstore needs `age`, so it is a real dependency of the repo,
not just of the author's machine). Tasks: `build`, `test`, `lint`, and
`setup`.

**`[hooks] postinstall = "mise run setup"`** — runs after `mise install`.
`setup` installs the git hooks and restores the packed agent context.

**`lefthook.yml`** — one `pre-commit` job, `mise run kstore:pack --stage`.
The pack task stages its own output, because lefthook's `stage_fixed` only
re-stages files the commit already had staged and these are artifacts the hook
itself creates.

**`.config/kstore` and `.config/kstore.sha256`** — the age-encrypted archive
and its digest stamp, both tracked. Verified to contain `CLAUDE.md` and
`.claude/rules/`, with `.claude/worktrees/` excluded.

## Restore is non-destructive

`mise run setup` unpacks only when `.config/kstore` exists **and** `CLAUDE.md`
does not. On a checkout that already has agent context it prints what it
skipped and how to force it, rather than overwriting local notes with an older
archive. A fresh clone gets `mise install` and has its rules back; nobody gets
their working notes clobbered by a hook they did not think about.

## Verified, not assumed

- `[hooks] postinstall` fires on `mise install` — tested in a scratch
  directory before being relied on here.
- The archive decrypts to exactly the expected paths.
- `lefthook install` writes to the shared `.git/hooks`, so the hook applies to
  the main checkout and every worktree.
- The full chain ran end to end by accident and worked: installing the missing
  `age` tool fired `postinstall`, which installed the hooks and restored
  `CLAUDE.md` and `.claude/rules/` into a worktree that had neither.

## Caveat: packing from a worktree

Every worktree now restores its own copy of the agent context, which is what
makes the rules present for agents working there. The consequence is that
`pack` run from a worktree packs *that* copy.

If `CLAUDE.md` is edited in the main checkout and then a commit is made from a
worktree holding an older copy, the archive is rewritten from the older one.
The digest stamp makes this visible — `.config/kstore` shows up modified in a
commit that had no reason to touch it — but nothing prevents it.

Edit agent context in the main checkout, and re-run `mise run setup` in a
worktree after doing so.

## Not added

The `pre-commit` hook runs `gofmt` and `go vet` in most repos. It does not
here, because there is no Go code yet and `go vet ./...` on an empty module
fails, which would block every commit. Those jobs get added with slice 0.
