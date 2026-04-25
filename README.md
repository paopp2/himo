# himo

A tiny file-based todo app. Plain markdown, vim-style TUI.

*Himo* is Cebuano for "to make" or "to do."

## Why

Most todo tools ask you to adopt their database, their query language, or their ritual. I wanted something that gets out of the way.

- **Simple workflow.** [Taskwarrior](https://taskwarrior.org/) and [dstask](https://github.com/naggie/dstask) are excellent but more than I need. I just want to write things down and cross them off.
- **File-based.** Your tasks are files on your disk. Back them up with the tool you already use (git, Dropbox, rsync, whatever). Delete `himo` and your todos are still valid markdown.
- **Markdown.** Open any project in your editor and it reads like a normal checklist. GitHub and your IDE render it correctly. No export, no import, no lock-in.

## Install

```sh
go install github.com/paopp2/himo/cmd/himo@latest
```

## Quick start

```sh
himo new work               # create a project
himo add "Write design doc" # capture a task from the shell
himo                        # open the TUI
```

First run prompts for a base directory (default `~/todos`).

## How it works

Each project is a directory with three files:

```
~/todos/
  work/
    active.md    # pending, in-progress, and blocked tasks
    done.md     # finished or cancelled, grouped by date
    backlog.md  # parked ideas
```

Tasks are plain markdown checkboxes:

```markdown
- [/] Write design doc
    Due Friday. Talk to Sam first.
- [!] Migrate the payments table
    Waiting on ops.
- [ ] Buy groceries
```

The checkbox marker is the status:

| Marker  | Status    |
|---------|-----------|
| `- [ ]` | pending   |
| `- [/]` | active    |
| `- [!]` | blocked   |
| `- [x]` | done      |
| `- [-]` | cancelled |
| `-`     | backlog   |

Indented content under a task is its notes. Notes stay with their task through every move.

Edit any file directly in `$EDITOR`. When you save, `himo` figures out what changed and moves tasks between files accordingly. Mark something done in `active.md` and it lands in `done.md` under today's date. Uncheck something in `done.md` and it goes back to `active.md`.

## TUI

Vim-style. Left pane is the task list, right pane previews the highlighted task's notes.

**Navigation:** `j`/`k` move, `g`/`G` top/bottom, `/` search, `q` quit

**Filters:** `0` all, `1` backlog, `2` pending, `3` active, `4` blocked, `5` done, `6` cancelled, `Esc` default

**Actions:** `Enter` open notes in editor, `e` edit title inline, `Space` cycle status, `o`/`O` new task, `d` delete, `!` block, `x` done, `-` cancel

**Scope:** `Tab`/`Shift+Tab` switch project, `P` project picker, `A` all projects

## Config

`~/.config/himo/config.toml`:

```toml
base_dir = "~/todos"
editor = "nvim"            # optional, falls back to $EDITOR
default_project = "work"   # optional
preview_pane = true        # optional
```

`HIMO_DIR` env var overrides `base_dir`. `$EDITOR` overrides `editor`.

## CLI

| Command                               | What it does                           |
|---------------------------------------|----------------------------------------|
| `himo`                                 | Open the TUI on the default project    |
| `himo <project>`                       | Open the TUI on a specific project     |
| `himo new <project>`                   | Create a new project                   |
| `himo add "<title>"`                   | Capture a pending task and exit        |
| `himo add -p <project> "<title>"`      | Capture to a specific project          |
| `himo ls [-p project] [-s status]`     | List tasks (scriptable)                |

## License

MIT. See [LICENSE](LICENSE).
