<br />
<p align="center">
  <img width="504" height="96" alt="diffnav" src="https://github.com/user-attachments/assets/b932225f-7f49-4274-886d-61e640f4ef8b" />
</p>

<p align="center">
  A git diff pager based on <a href="https://github.com/dandavison/delta">delta</a> but with a file tree, à la GitHub.
  <br />
  <br />
  <em>Mathew's personal fork of <a href="https://github.com/dlvhdr/diffnav">dlvhdr/diffnav</a>, focused on speeding up local code review in the AI age. See <a href="./IDEAS.md">IDEAS.md</a> for the roadmap.</em>
</p>

<p align="center">
  <img width="900" src="https://github.com/user-attachments/assets/104e156e-7e9d-4ea5-bea1-399ca71e12a5" />
</p>

## Installation

This fork is not published to Homebrew. Install from source:

```sh
go install github.com/matart/diffnav@latest
```

Or clone and build:

```sh
git clone https://github.com/matart/diffnav.git
cd diffnav
go install .
```

> [!NOTE]
> To get the icons to render properly you should download and install a Nerd font from https://www.nerdfonts.com/. Then, select that font as your font for the terminal.
>
> _You can install these with brew as well: `brew install --cask font-<FONT NAME>-nerd-font`_

## Usage

### Pipe into `diffnav`

- `git diff | diffnav`
- `gh pr diff <PR URL> | diffnav`

### Set up as Global Git Diff Pager

```bash
git config --global pager.diff diffnav
```

## Flags

| Flag                 | Description                                                           |
| -------------------- | --------------------------------------------------------------------- |
| `--side-by-side, -s` | Force side-by-side diff view                                          |
| `--unified, -u`      | Force unified diff view                                               |
| `--watch, -w`        | Watch mode: periodically re-run a command and refresh                 |
| `--watch-cmd`        | Command to run in watch mode (implies `--watch`, default: `git diff`) |
| `--watch-interval`   | Interval between watch refreshes (default: `2s`)                      |

Example:

```sh
git diff | diffnav --unified
git diff | diffnav -u
```

### Watch Mode

Watch mode lets diffnav periodically re-run a diff command and refresh the display automatically. This is useful for monitoring changes as you work.

```sh
# watch unstaged changes (default: git diff, every 2s)
diffnav --watch

# watch staged changes with a custom interval
diffnav --watch-cmd "git diff --cached" --watch-interval 5s

# watch changes against a specific branch
diffnav --watch-cmd "git diff main..."
```

## Configuration

The config file is searched in this order:

1. `$DIFFNAV_CONFIG_DIR/config.yml` (if env var is set)
2. `$XDG_CONFIG_HOME/diffnav/config.yml` (if set, macOS only)
3. `~/.config/diffnav/config.yml` (macOS and Linux)
4. OS-specific config directory (e.g., `~/Library/Application Support/diffnav/config.yml` on macOS)

Example config file:

```yaml
ui:
  # Hide the header to get more screen space for diffs
  hideHeader: true

  # Hide the footer (keybindings help)
  hideFooter: true

  # Start with the file tree hidden (toggle with 'e')
  showFileTree: false

  # Customize the file tree width (default: 26)
  fileTreeWidth: 30

  # Customize the search panel width (default: 50)
  searchTreeWidth: 60

  # Icon style: "status" (default), "simple", "filetype", "full", "unicode", or "ascii"
  icons: nerd-fonts-status

  # Color filenames by git status (default: true)
  colorFileNames: false

  # Show the amount of lines added / removed next to the file
  showDiffStats: false

  # Use side-by-side diff view (default: true, set false for unified)
  sideBySide: true

  # How many levels of folders to open on start (-1 = all, 0 = none, 1 = first level, etc.)
  startFoldersOpenDepth: 1
```

| Option                     | Type   | Default             | Description                                               |
| :------------------------- | :----- | :------------------ | :-------------------------------------------------------- |
| `ui.hideHeader`            | bool   | `false`             | Hide the "DIFFNAV" header                                 |
| `ui.hideFooter`            | bool   | `false`             | Hide the footer with keybindings help                     |
| `ui.showFileTree`          | bool   | `true`              | Show file tree on startup                                 |
| `ui.fileTreeWidth`         | int    | `26`                | Width of the file tree sidebar                            |
| `ui.searchTreeWidth`       | int    | `50`                | Width of the search panel                                 |
| `ui.icons`                 | string | `nerd-fonts-status` | Icon style (see below for details)                        |
| `ui.colorFileNames`        | bool   | `true`              | Color filenames by git status                             |
| `ui.showDiffStats`         | bool   | `true`              | Show the amount of lines added / removed next to the file |
| `ui.sideBySide`            | bool   | `true`              | Use side-by-side diff view (false for unified)            |
| `ui.startFoldersOpenDepth` | int    | `-1`                | Folder open depth on start (-1 = all, 0 = none)           |

### Icon Styles

| Style                 | Description                                                      |
| :-------------------- | :--------------------------------------------------------------- |
| `nerd-fonts-status`   | Boxed git status icons colored by change type                    |
| `nerd-fonts-simple`   | Generic file icon colored by change type                         |
| `nerd-fonts-filetype` | File-type specific icons (language icons) colored by change type |
| `nerd-fonts-full`     | Both status icon and file-type icon, all colored                 |
| `unicode`             | Unicode symbols (+/⛌/●)                                          |
| `ascii`               | Plain ASCII characters (+/x/\*)                                  |

### Storage

`diffnav` persists reviewed-hunk markers (toggled with <kbd>r</kbd>) to a JSON file.
By default it's `~/.local/share/diffnav/reviewed.json` (or `$XDG_DATA_HOME/diffnav/`
if set). Override via env var or config:

```yaml
storage:
  path: ~/Library/Application Support/diffnav
```

Env var override: `DIFFNAV_STORAGE_DIR` (priority: env > yaml > default).

### Delta

You can also configure the diff rendering through delta. Check out [their docs](https://dandavison.github.io/delta/configuration.html).

This fork's delta config lives at [`cfg/delta.conf`](./cfg/delta.conf).

## Keys

| Key                         | Description                      |
| :-------------------------- | :------------------------------- |
| <kbd>j</kbd>                | Next node                        |
| <kbd>k</kbd>                | Previous node                    |
| <kbd>n</kbd>                | Next file                        |
| <kbd>p</kbd> / <kbd>N</kbd> | Previous file                    |
| <kbd>Ctrl-d</kbd>           | Scroll the diff down             |
| <kbd>Ctrl-u</kbd>           | Scroll the diff up               |
| <kbd>e</kbd>                | Toggle the file tree             |
| <kbd>t</kbd>                | Search/go-to file                |
| <kbd>y</kbd>                | Copy file path                   |
| <kbd>i</kbd>                | Cycle icon style                 |
| <kbd>o</kbd>                | Open file in $EDITOR             |
| <kbd>r</kbd>                | Toggle current hunk as reviewed  |
| <kbd>]</kbd> / <kbd>[</kbd> | Next / previous hunk             |
| <kbd>s</kbd>                | Toggle side-by-side/unified view |
| <kbd>Tab</kbd>              | Switch focus between the panes   |
| <kbd>q</kbd>                | Quit                             |

## Credits

This is a fork of [dlvhdr/diffnav](https://github.com/dlvhdr/diffnav). All credit for the original tool goes to [@dlvhdr](https://github.com/dlvhdr) and the upstream contributors. If you want to support the upstream project, see the [sponsors page](https://github.com/sponsors/dlvhdr).

## Under the Hood

`diffnav` uses:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- [`delta`](https://github.com/dandavison/delta) for viewing the diffed file
