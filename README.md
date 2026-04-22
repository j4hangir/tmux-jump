# tmux-jump

Incremental-narrowing jump for tmux — like neovim's flash / leap, scoped to the visible pane.

Press the jump key, type characters of a visible word, and the moment only one match remains the copy-mode cursor lands on it.

## How it works

1. Capture the visible pane (`tmux capture-pane -p`).
2. Open a borderless `display-popup` sized to exactly overlay the pane.
3. Inside the popup a Go TUI renders the captured content with matches highlighted (black on yellow) and everything else dimmed.
4. Each keystroke narrows the match set using **smart-case substring** search (query has an uppercase letter → case-sensitive; else case-insensitive).
5. When one match remains, the popup closes and the original pane enters copy-mode with the cursor placed on that (row, column).

## Install

### With TPM

```tmux
set -g @plugin 'j4hangir/tmux-jump'
```

### Manual

```sh
git clone https://git.j4hangir.com/tmux/tmux-jump ~/.tmux/plugins/tmux-jump
echo "run-shell ~/.tmux/plugins/tmux-jump/tmux-jump.tmux" >> ~/.tmux.conf
tmux source-file ~/.tmux.conf
```

On first launch, a menu offers to download the prebuilt Linux x86_64 binary from GitLab CI artifacts or build from source (Go 1.22+ required).

## Build

Requires Go 1.22+ and tmux ≥ 3.2 (for `display-popup -B`).

```sh
make build
make test
```

## Keys (inside jump mode)

| Key | Action |
| --- | --- |
| printable | append to query, re-narrow |
| Backspace | pop last character |
| Enter | jump to first match |
| Esc / Ctrl-C / Ctrl-G | cancel |

Unique match → auto-jump, no Enter required.
Zero matches after a keystroke → bell, character rejected.

## Config

```tmux
set -g @jump-key j   # default: j (invoked as prefix + j)
```

## Limitations (v1)

- ASCII-first column math. Lines with CJK / wide characters may jump to the wrong column.
- Works on the active pane only.
- Requires tmux ≥ 3.2 for borderless pane-sized popups.

## Credits

Inspired by `schasse/tmux-jump` (original hint-based approach), `fcsonline/tmux-fingers` (overlay pattern), and `flash.nvim` / `leap.nvim` (incremental-narrowing UX).
