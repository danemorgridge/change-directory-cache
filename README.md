# cdc — change directory cache

A tiny cross-platform CLI to bookmark directories and jump into them. No shell
hooks, no PATH editing, no PowerShell — install and use.

## Install

Once it's published to GitHub:

```sh
go install github.com/danemorgridge/change-directory-cache/cmd/cdc@latest
```

Or install straight from a local checkout (no push required) — run from the repo root:

```sh
go install ./cmd/cdc
```

Either way, `go install` drops the `cdc` binary in your Go bin directory (already on
your PATH), so `cdc` works immediately in a new shell.

## Usage

```
cdc save my project     # bookmark the current folder as "my project"
cdc o proj              # open a shell in a bookmark whose name/path contains "proj"
cdc list                # show all bookmarks as a table
cdc rm                  # forget the current folder
cdc export backup.json  # write all bookmarks to a file (or stdout if omitted)
cdc import backup.json  # merge bookmarks from a file
```

| Command | Aliases | What it does |
|---|---|---|
| `cdc open <query>` | `o` | Case-insensitive substring match of `<query>` against every bookmark's **name or path**. One match opens it; several show an arrow-key picker; none is an error. |
| `cdc save [name...]` | `s` | Save the **current** directory. The name is all text after `s` (spaces allowed); with no name it defaults to the folder's basename. If the folder is already saved, it reports the existing name and asks whether to update it. |
| `cdc rm` | | Remove the **current** directory from the cache (errors if it isn't saved). |
| `cdc list [-p]` | `ls` | Print a `NAME  FOLDER` table sorted by name. `-p` sorts by full path instead. |
| `cdc export [file]` | | Write the whole cache as JSON to `file`. With no argument (or `-`) it prints to stdout, so it can be piped. |
| `cdc import <file>` | | Merge bookmarks from a JSON file (same shape as `export`). Entries whose **directory no longer exists** are reported and skipped; entries that **duplicate a path** already in the cache are reported and ignored. A leading UTF-8 BOM is tolerated. |

## How "jumping" works

A program can't change its parent shell's current directory — that's an OS-level
limitation, which is why tools like `zoxide` and `direnv` require you to install a
shell hook. `cdc` avoids all of that: `cdc open` **starts a new shell already sitting
in the target directory**. Type `exit` (or Ctrl-D) to return to where you were.

- On **Windows** it launches `cmd.exe` (via `%COMSPEC%`) — no PowerShell needed.
- On **macOS / Linux** it launches your `$SHELL`.
- Set the `CDC_SHELL` environment variable to force a specific shell, e.g.
  `CDC_SHELL=/usr/bin/fish` or `CDC_SHELL=pwsh`.

The bookmarks live in a JSON file at `~/.directory-cache.json`.

## Build from source

```sh
go build -o cdc ./cmd/cdc      # add .exe on Windows: go build -o cdc.exe ./cmd/cdc
```
