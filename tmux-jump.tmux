#!/usr/bin/env bash
# tmux-jump plugin entry: locates the binary (PATH, then ./bin),
# triggers the install wizard if missing, prompts to reinstall if
# the binary is older than VERSION, then binds the jump key.
#
# Defensive: every tmux call that could fail at plugin-load time is
# wrapped with `|| true` so a single error (e.g. an option unset) can
# never abort the tmux startup sequence mid-way.

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
JUMP_BINARY=""

if command -v tmux-jump >/dev/null 2>&1; then
	JUMP_BINARY="tmux-jump"
elif [[ -x "$CURRENT_DIR/bin/tmux-jump" ]]; then
	JUMP_BINARY="$CURRENT_DIR/bin/tmux-jump"
fi

if [[ -z "$JUMP_BINARY" ]]; then
	tmux run-shell -b "bash $CURRENT_DIR/install-wizard.sh" 2>/dev/null || true
	exit 0
fi

CURRENT_JUMP_VERSION=$("$JUMP_BINARY" version 2>/dev/null || echo "")
CURRENT_GIT_VERSION=$(cat "$CURRENT_DIR/VERSION" 2>/dev/null || echo "")
SKIP_WIZARD=$(tmux show-option -gqv @jump-skip-wizard 2>/dev/null || echo "0")
SKIP_WIZARD=${SKIP_WIZARD:-0}

if [[ "$SKIP_WIZARD" = "0" && -n "$CURRENT_GIT_VERSION" && "$CURRENT_JUMP_VERSION" != "$CURRENT_GIT_VERSION" ]]; then
	tmux run-shell -b "JUMP_UPDATE=1 bash $CURRENT_DIR/install-wizard.sh" 2>/dev/null || true
fi

key=$(tmux show-option -gqv @jump-key 2>/dev/null || echo "")
[[ -z "$key" ]] && key=j

tmux bind-key -N "Jump to visible text in copy mode" "$key" \
	run-shell -b "JUMP_BINARY='$JUMP_BINARY' $CURRENT_DIR/tmux-jump.sh" 2>/dev/null || true

exit 0
