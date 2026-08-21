# treex

Reclaim the disk space AI coding agents leave behind.

Agents work in throwaway worktrees and fill them with `node_modules`, `dist`,
`Pods`, and `DerivedData`. They cache toolchains, keep transcripts, and leave
sockets behind when they exit. None of it gets cleaned up, and the totals get
large: on the machine this was built for, `~/.codex` alone had grown to 236 GB.

`rm -rf` is the wrong tool for it, for two reasons that treex exists to encode.

**Agent worktrees are usually _linked_ git worktrees.** Their `.git` is a file
pointing back into a parent repository that holds a registry entry for them.
Deleting the directory leaves that entry dangling — the state git calls
`prunable` — and those accumulate silently. The first `treex doctor` run on the
machine above found 21 of them across five repositories.

**The rest are clones, and clones hold work.** One of them had four modified
files and no upstream; another had 111 commits that existed on no remote. Both
looked exactly like the ones that were safe to delete.

treex measures first, refuses to delete anything holding work, and removes
linked worktrees by asking git to do it.

## Install

```bash
brew install --cask treex
```

Or run it from a container against a mounted home directory:

```bash
docker run --rm -v "$HOME:/host:ro" ghcr.io/oullin/treex scan --root /host
```

## Use

Start by looking. `scan` never modifies anything:

```bash
treex scan
treex scan --providers codex --explain
```

`clean` is a dry run unless you pass `--apply`. There is no `--dry-run` flag,
because not deleting is the default:

```bash
treex clean                    # print the plan, change nothing
treex clean --apply            # prompt, then act
treex clean --apply --yes      # unattended; required outside a terminal
```

`doctor` explains what a scan could not act on, which is the usual first
question:

```bash
treex doctor                   # report stale registries and blocked worktrees
treex doctor --apply           # repair the registries
```

### Where to start

`--categories artifact` is the safe daily driver. It removes build output from
*inside* worktrees it keeps, so it carries no git risk at all — including for
worktrees treex has refused to delete:

```bash
treex clean --categories artifact --apply
```

## What it sweeps

| Category   | What it is                                        | Default |
|------------|---------------------------------------------------|---------|
| `worktree` | A whole agent worktree or clone                   | on      |
| `artifact` | `node_modules`, `dist`, `build`, … inside one     | on      |
| `cache`    | Agent scratch and toolchain caches                | on      |
| `orphan`   | Files whose owning process has exited             | on      |
| `session`  | Transcripts, chat history, logs                   | **off** |

Sessions are the one class that cannot be regenerated, so they are only ever
swept when named explicitly.

## Safety

A candidate is refused if it has uncommitted changes, untracked files, commits
on no remote, a stash of its own, or a protected branch checked out. `--force`
relaxes those, because they are yours to waive.

`--force` does not relax the structural guards, which is the point of keeping
them separate:

- the primary working tree of a repository with worktrees hanging off it
- anything inside a protected path (`~/Sites`, `~/Documents`, … by default)
- anything outside every configured provider root
- a symlink, or a path whose inode changed since it was measured

Untracked build directories are ignored when judging whether a worktree holds
work. Without that, the untracked `node_modules` in every agent worktree would
make every one of them look dirty and treex would reclaim nothing.

Linked worktrees are always removed through `git worktree remove`. If git
refuses, treex reports it and moves on — falling back to unlinking the directory
would create exactly the corruption this tool cleans up.

## Configuring

treex ships knowing about `claude`, `codex`, `cursor`, `gemini`, `grok`, and
`agent-browser`, with `copilot`, `factory`, `pi`, and the toolchain caches
available but off.

```bash
treex providers                # what treex knows and where it resolves to
treex config init              # write a starter file
treex config path
```

Configuration lives at `~/.config/treex/config.yml` and *patches* the built-in
catalog rather than replacing it, so most files are a few lines:

```yaml
version: 1

providers:
    - name: cursor
      enabled: false

    - name: codex
      caches: [cache, runtimes, android, maestro, my-extra-cache]
```

Adding a tool treex has never heard of works the same way:

```yaml
providers:
    - name: my-agent
      root: ~/.my-agent
      worktrees:
          - { path: worktrees, depth: 2 }
      caches: [cache]
```

## Layout

| Package             | What it does                                            |
|---------------------|---------------------------------------------------------|
| `config`            | The YAML schema and the built-in provider catalog        |
| `report`            | The output contract, text and JSON                       |
| `internal/diskwalk` | Concurrent tree sizing                                   |
| `internal/gitwork`  | Worktree classification and the removal policy           |
| `internal/inventory`| Finding candidates and measuring them                    |
| `internal/plan`     | Turning an inventory into ordered actions                |
| `internal/sweep`    | The only code that deletes, and the guard in front of it |
| `internal/app`      | The composition root and the command surface             |

`config` and `report` sit outside `internal` because they are a public
contract: the file users write and the JSON other tools parse.

## Developing

```bash
pnpm exec vp run go:test        # the whole repository, treex included
pnpm exec vp run treex:run -- scan
pnpm exec vp run format
```

Releases are tag-driven: push a `treex/vX.Y.Z` tag, or run the `treex Release`
workflow by hand.
