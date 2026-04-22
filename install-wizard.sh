#!/usr/bin/env bash
# tmux-jump install wizard: offers a menu to download a prebuilt Linux binary
# from the GitLab project's artifacts, or to build from source with Go.
# Adapted from tmux-fingers/install-wizard.sh.

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PLATFORM=$(uname -s)
ARCH=$(uname -m)
VERSION=$(cat "$CURRENT_DIR/VERSION" 2>/dev/null || echo "?")
action=$1

PROJECT_API="https://git.j4hangir.com/api/v4/projects/tmux%2Ftmux-jump"
PROJECT_URL="https://git.j4hangir.com/tmux/tmux-jump"

function finish {
	exit_code=$?

	if [[ -z "$action" ]]; then
		exit $exit_code
	fi

	if [[ $exit_code -eq 0 ]]; then
		echo "Reloading tmux.conf"
		tmux source ~/.tmux.conf 2>/dev/null || true
		exit 0
	else
		echo "Something went wrong. Press any key to close this window"
		read -n 1
		exit 1
	fi
}
trap finish EXIT

function install_from_source() {
	echo "Installing tmux-jump v$VERSION from source..."

	if ! command -v go >/dev/null 2>&1; then
		echo "Go is not installed. Please install it first:"
		echo ""
		echo "  https://go.dev/dl/"
		echo ""
		exit 1
	fi

	pushd "$CURRENT_DIR" > /dev/null
		mkdir -p bin
		CGO_ENABLED=0 go build -ldflags "-X main.version=$VERSION" -o bin/tmux-jump .
	popd > /dev/null

	echo "Build complete!"
	exit 0
}

function download_binary() {
	mkdir -p "$CURRENT_DIR/bin"

	if [[ "$ARCH" != "x86_64" ]]; then
		echo "tmux-jump binaries are only provided for x86_64. Build from source instead."
		exit 1
	fi
	if [[ "$PLATFORM" != "Linux" ]]; then
		echo "tmux-jump binaries are only built for Linux. Build from source on $PLATFORM."
		exit 1
	fi

	echo "Getting latest tag..."
	tags=$(curl -sSf "$PROJECT_API/repository/tags" 2>&1)
	if [[ $? -ne 0 || -z "$tags" || "$tags" == "[]" ]]; then
		echo "Could not fetch tags from $PROJECT_API/repository/tags"
		echo "Response: $tags"
		exit 1
	fi
	tag=$(echo "$tags" | grep -o '"name":"[^"]*"' | head -1 | sed 's/"name":"\([^"]*\)"/\1/')
	if [[ -z "$tag" ]]; then
		echo "Could not extract tag name from tags response."
		echo "Response: $tags"
		exit 1
	fi

	url="$PROJECT_URL/-/jobs/artifacts/$tag/raw/tmux-jump?job=build"
	echo "Installing tmux-jump v$VERSION (binary: $tag)..."

	if ! curl -sSfL "$url" -o "$CURRENT_DIR/bin/tmux-jump"; then
		echo "Failed to download binary. The CI build for $tag may still be running or may have failed."
		echo "Check $PROJECT_URL/-/pipelines"
		exit 1
	fi
	chmod a+x "$CURRENT_DIR/bin/tmux-jump"

	echo "Download complete!"
	exit 0
}

if [[ "$action" == "download-binary" ]]; then
	download_binary
fi

if [[ "$action" == "install-from-source" ]]; then
	install_from_source
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
	if [[ "$JUMP_UPDATE" == "1" ]]; then
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
	"Exit" q ""
