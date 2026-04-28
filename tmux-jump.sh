#!/usr/bin/env bash
# tmux-jump wrapper: captures the visible pane, runs the Go TUI in a
# borderless popup sized to the pane, then moves the copy-mode cursor
# to the (row, col) picked by the user.
#
# Defensive: every failure path falls back to a no-op cleanly. The
# wrapper never leaves the pane stuck in copy-mode at a half-moved
# position — if we entered copy-mode but didn't complete the jump,
# the cleanup trap cancels copy-mode for us.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${JUMP_BINARY:-$SCRIPT_DIR/bin/tmux-jump}"

if ! command -v "$BIN" >/dev/null 2>&1 && [ ! -x "$BIN" ]; then
	tmux run-shell -b "bash $SCRIPT_DIR/install-wizard.sh" 2>/dev/null || true
	exit 0
fi

tmpdir=$(mktemp -d 2>/dev/null) || exit 0
pane=""
entered_copy=0
completed=0

cleanup() {
	rm -rf "$tmpdir" 2>/dev/null || true
	if [ "$entered_copy" = 1 ] && [ "$completed" = 0 ] && [ -n "$pane" ]; then
		tmux send-keys -t "$pane" -X cancel 2>/dev/null || true
	fi
}
trap cleanup EXIT

info=$(tmux display-message -p -F '#{pane_id} #{pane_width} #{pane_height} #{pane_in_mode} #{pane_mode} #{scroll_position}' 2>/dev/null) || exit 0
read -r pane w h in_mode mode scroll_pos <<<"$info"
[ -n "${pane:-}" ] || exit 0

in_copy=0
if [ "${in_mode:-0}" = 1 ] && [ "${mode:-}" = "copy-mode" ]; then
	in_copy=1
fi
case "${scroll_pos:-}" in '' | *[!0-9]*) scroll_pos=0 ;; esac

cap="$tmpdir/cap"
res="$tmpdir/res"

if [ "$in_copy" = 1 ] && [ "$scroll_pos" -gt 0 ]; then
	cap_start=$((-scroll_pos))
	cap_end=$((-scroll_pos + h - 1))
	tmux capture-pane -p -t "$pane" -S "$cap_start" -E "$cap_end" >"$cap" 2>/dev/null || exit 0
else
	tmux capture-pane -p -t "$pane" >"$cap" 2>/dev/null || exit 0
fi

hints_arg=""
if [ -n "${JUMP_HINTS:-}" ]; then
	hints_arg="-hints $(printf %q "$JUMP_HINTS")"
fi

tmux display-popup -E -B -w "$w" -h "$h" -x P -y P \
	"$BIN -capture $(printf %q "$cap") -out $(printf %q "$res") -w $w -h $h $hints_arg" \
	2>/dev/null || exit 0

[ -s "$res" ] || exit 0

IFS=, read -r row col len <"$res" || exit 0
[ -n "${row:-}" ] && [ -n "${col:-}" ] || exit 0
case "$row" in '' | *[!0-9]*) exit 0 ;; esac
case "$col" in '' | *[!0-9]*) exit 0 ;; esac
case "${len:-}" in *[!0-9]*) len="" ;; esac

if [ "$in_copy" != 1 ]; then
	tmux copy-mode -t "$pane" 2>/dev/null || exit 0
	entered_copy=1
fi
tmux send-keys -t "$pane" -X top-line 2>/dev/null || exit 0
tmux send-keys -t "$pane" -X start-of-line 2>/dev/null || exit 0
if [ "$row" -gt 0 ]; then
	tmux send-keys -t "$pane" -N "$row" -X cursor-down 2>/dev/null || exit 0
fi
if [ "$col" -gt 0 ]; then
	tmux send-keys -t "$pane" -N "$col" -X cursor-right 2>/dev/null || exit 0
fi
if [ "${JUMP_SELECT:-0}" = 1 ] && [ -n "${len:-}" ] && [ "$len" -gt 0 ]; then
	tmux send-keys -t "$pane" -X begin-selection 2>/dev/null || exit 0
	if [ "$len" -gt 1 ]; then
		tmux send-keys -t "$pane" -N "$((len - 1))" -X cursor-right 2>/dev/null || exit 0
	fi
fi
completed=1
exit 0
