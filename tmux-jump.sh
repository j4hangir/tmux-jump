#!/usr/bin/env bash
# tmux-jump wrapper: captures the visible pane, runs the Go TUI in a
# borderless popup sized to the pane, then moves the copy-mode cursor
# to the (row, col) picked by the user.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${JUMP_BINARY:-$SCRIPT_DIR/bin/tmux-jump}"

if ! command -v "$BIN" >/dev/null 2>&1 && [ ! -x "$BIN" ]; then
	tmux run-shell -b "bash $SCRIPT_DIR/install-wizard.sh"
	exit 1
fi

tmpdir=$(mktemp -d)
cap="$tmpdir/cap"
res="$tmpdir/res"
trap 'rm -rf "$tmpdir"' EXIT

eval "$(tmux display-message -p 'pane=#{pane_id}; w=#{pane_width}; h=#{pane_height}; x=#{pane_left}; y=#{pane_top}')"

tmux capture-pane -p -t "$pane" > "$cap"

tmux display-popup -E -B -w "$w" -h "$h" -x "$x" -y "$y" \
	"$BIN -capture $(printf %q "$cap") -out $(printf %q "$res") -w $w -h $h"

[ -s "$res" ] || exit 0

IFS=, read -r row col < "$res"
[ -n "${row:-}" ] || exit 0
[ -n "${col:-}" ] || exit 0

tmux copy-mode -t "$pane"
tmux send-keys -t "$pane" -X top-line
tmux send-keys -t "$pane" -X start-of-line
[ "$row" -gt 0 ] && tmux send-keys -t "$pane" -N "$row" -X cursor-down
[ "$col" -gt 0 ] && tmux send-keys -t "$pane" -N "$col" -X cursor-right
exit 0
