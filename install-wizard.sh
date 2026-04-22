#!/usr/bin/env bash
# tmux-jump install wizard: offers a menu to download a prebuilt Linux binary
# from the GitLab project's artifacts, or to build from source with Go.
# Adapted from tmux-fingers/install-wizard.sh.
#
# All failure paths yield control back cleanly: the trap always runs,
# curl has a timeout, and partial downloads are staged via a .tmp file
# so a broken binary never ends up at bin/tmux-jump.

set -eo pipefail

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PLATFORM=$(uname -s 2>/dev/null || echo unknown)
ARCH=$(uname -m 2>/dev/null || echo unknown)
VERSION=$(cat "$CURRENT_DIR/VERSION" 2>/dev/null || echo "?")
action=${1:-}

PROJECT_API="https://git.j4hangir.com/api/v4/projects/tmux%2Ftmux-jump"
PROJECT_URL="https://git.j4hangir.com/tmux/tmux-jump"

function finish {
	local exit_code=$?

	if [[ -z "$action" ]]; then
		exit $exit_code
	fi

	if [[ $exit_code -eq 0 ]]; then
		echo "Reloading tmux.conf"
		tmux source ~/.tmux.conf 2>/dev/null || true
		exit 0
	else
		echo "Something went wrong. Press any key to close this window"
		read -n 1 -r _ || true
		exit 1
	fi
}
trap finish EXIT

function install_from_source() {
	echo "Installing tmux-jump v$VERSION from source..."
	if ! command -v go >/dev/null 2>&1; then
		echo "Go is not installed. Install from https://go.dev/dl/"
		return 1
	fi
	cd "$CURRENT_DIR" || return 1
	mkdir -p bin
	if ! CGO_ENABLED=0 go build -ldflags "-X main.version=$VERSION" -o bin/tmux-jump . ; then
		echo "go build failed"
		return 1
	fi
	echo "Build complete!"
	return 0
}

function download_binary() {
	mkdir -p "$CURRENT_DIR/bin"

	if [[ "$ARCH" != "x86_64" ]]; then
		echo "tmux-jump binaries are only published for x86_64. Build from source instead."
		return 1
	fi
	if [[ "$PLATFORM" != "Linux" ]]; then
		echo "tmux-jump binaries are only built for Linux. Build from source on $PLATFORM."
		return 1
	fi

	echo "Getting latest tag..."
	local tags
	if ! tags=$(curl -sSf --max-time 15 "$PROJECT_API/repository/tags" 2>&1); then
		echo "Could not fetch tags from $PROJECT_API/repository/tags"
		echo "Response: $tags"
		return 1
	fi
	if [[ -z "$tags" || "$tags" == "[]" ]]; then
		echo "No tags found — CI hasn't published a release yet."
		echo "Build from source, or wait for a tag at $PROJECT_URL/-/tags"
		return 1
	fi

	local tag
	tag=$(printf '%s' "$tags" | grep -o '"name":"[^"]*"' | head -1 | sed 's/"name":"\([^"]*\)"/\1/')
	if [[ -z "$tag" ]]; then
		echo "Could not extract tag name from tags response."
		return 1
	fi

	local url="$PROJECT_URL/-/jobs/artifacts/$tag/raw/tmux-jump?job=build"
	echo "Installing tmux-jump v$VERSION (binary: $tag)..."

	# Download to a .tmp file; atomic rename on success. This ensures
	# we never leave a corrupt binary at bin/tmux-jump if curl dies.
	local tmp="$CURRENT_DIR/bin/tmux-jump.tmp"
	if ! curl -sSfL --max-time 120 "$url" -o "$tmp"; then
		rm -f "$tmp"
		echo "Failed to download binary. The CI build for $tag may still be running or may have failed."
		echo "Check $PROJECT_URL/-/pipelines"
		return 1
	fi
	if [[ ! -s "$tmp" ]]; then
		rm -f "$tmp"
		echo "Downloaded artifact is empty."
		return 1
	fi
	mv "$tmp" "$CURRENT_DIR/bin/tmux-jump"
	chmod a+x "$CURRENT_DIR/bin/tmux-jump"
	echo "Download complete!"
	return 0
}

if [[ "$action" == "download-binary" ]]; then
	download_binary
	exit $?
fi

if [[ "$action" == "install-from-source" ]]; then
	install_from_source
	exit $?
fi

function default_install_label() {
	if [[ "$PLATFORM" == "Linux" && "$ARCH" == "x86_64" ]]; then
		echo "Download binary"
	else
		echo "Build from source (Go required)"
	fi
}

function default_install_action() {
	if [[ "$PLATFORM" == "Linux" && "$ARCH" == "x86_64" ]]; then
		echo "download-binary"
	else
		echo "install-from-source"
	fi
}

function get_message() {
	if [[ "${JUMP_UPDATE:-}" == "1" ]]; then
		echo "tmux-jump has been updated. We need to rebuild or redownload the binary."
	else
		echo "First run — we need to install the tmux-jump binary before the key binding works."
	fi
}

tmux display-menu -T "tmux-jump v$VERSION" \
	"" \
	"- " "" "" \
	"-  #[nodim,bold]Welcome to tmux-jump 🐇" "" "" \
	"- " "" "" \
	"-  $(get_message) " "" "" \
	"- " "" "" \
	"" \
	"$(default_install_label)" b "new-window \"$CURRENT_DIR/install-wizard.sh $(default_install_action)\"" \
	"Build from source (Go required)" s "new-window \"$CURRENT_DIR/install-wizard.sh install-from-source\"" \
	"" \
	"Exit" q "" 2>/dev/null || true
