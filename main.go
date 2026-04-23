package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	capturePath = flag.String("capture", "", "path to captured pane file")
	outPath     = flag.String("out", "", "path to write result (row,col)")
	width       = flag.Int("w", 80, "overlay width")
	height      = flag.Int("h", 24, "overlay height")
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Println(version)
		return
	}
	os.Exit(run())
}

// run does all work inside a function so deferred tty/stty restores
// always fire — even on error paths. Never call os.Exit here.
func run() (code int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "tmux-jump panic:", r)
			code = 1
		}
	}()

	flag.Parse()
	if *capturePath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: tmux-jump -capture FILE -out FILE [-w W] [-h H]")
		return 2
	}

	data, err := os.ReadFile(*capturePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	rows := parseRows(string(data))

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer tty.Close()

	saved, err := sttySave(tty)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer sttyRestore(tty, saved)

	if err := sttyRaw(tty); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer fmt.Fprint(tty, ansiShowCur+ansiReset)

	query := []rune{}
	var matches []Match
	selected := 0

	for {
		fmt.Fprint(tty, render(rows, matches, selected, string(query), *width, *height))

		b, ok := readByte(tty, true)
		if !ok {
			return 0
		}

		// Escape sequence: ESC [ ... or ESC O ...
		if b == 0x1b {
			b2, ok2 := readByte(tty, false)
			if !ok2 {
				return 0 // bare Esc
			}
			if b2 != '[' && b2 != 'O' {
				continue
			}
			b3, ok3 := readByte(tty, false)
			if !ok3 {
				continue
			}
			switch b3 {
			case 'A', 'D': // up, left
				if n := len(matches); n > 1 && n <= selectLimit {
					selected = (selected - 1 + n) % n
				}
			case 'B', 'C', 'Z': // down, right, shift-tab
				if n := len(matches); n > 1 && n <= selectLimit {
					selected = (selected + 1) % n
				}
			}
			continue
		}

		switch {
		case b == 0x03, b == 0x07:
			return 0
		case b == 0x09: // Tab → next
			if n := len(matches); n > 1 && n <= selectLimit {
				selected = (selected + 1) % n
			}
		case b == 0x7f, b == 0x08:
			if len(query) > 0 {
				query = query[:len(query)-1]
				matches = findMatches(rows, query)
				selected = defaultSelected(len(matches))
			}
		case b == '\r', b == '\n':
			if len(matches) >= 1 && selected < len(matches) {
				writeResult(matches[selected])
				return 0
			}
		case b >= 0x20 && b < 0x7f:
			newQ := append(append([]rune{}, query...), rune(b))
			newM := findMatches(rows, newQ)
			if len(newM) == 0 {
				fmt.Fprint(tty, "\x07")
				continue
			}
			query = newQ
			matches = newM
			selected = defaultSelected(len(matches))
			if len(matches) == 1 {
				writeResult(matches[0])
				return 0
			}
		}
	}
}

const selectLimit = 9

// defaultSelected picks the LAST match when the set is small enough
// to navigate; otherwise 0 (unused until navigation engages).
func defaultSelected(n int) int {
	if n > 0 && n <= selectLimit {
		return n - 1
	}
	return 0
}

// readByte reads one byte. If blocking is true, retries on timeout
// (stty is set to min 0 time 1). If false, returns (0,false) on timeout,
// used for peeking escape-sequence continuation bytes.
//
// On Linux tty with VMIN=0, read(2) returning 0 bytes means "timeout, no
// data" — but Go's *os.File surfaces that as io.EOF because ZeroReadIsEOF
// is set for file-kind descriptors. Treat EOF (and any zero-byte read) as
// a timeout, not a fatal error.
func readByte(tty *os.File, blocking bool) (byte, bool) {
	buf := make([]byte, 1)
	for {
		n, err := tty.Read(buf)
		if n == 1 {
			return buf[0], true
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, false
		}
		if !blocking {
			return 0, false
		}
	}
}

func parseRows(s string) [][]rune {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	out := make([][]rune, len(lines))
	for i, l := range lines {
		out[i] = []rune(l)
	}
	return out
}

func writeResult(m Match) {
	_ = os.WriteFile(*outPath, []byte(fmt.Sprintf("%d,%d\n", m.Row, m.Col)), 0644)
}

func sttySave(tty *os.File) (string, error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = tty
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func sttyRaw(tty *os.File) error {
	cmd := exec.Command("stty", "-icanon", "-echo", "min", "0", "time", "1")
	cmd.Stdin = tty
	return cmd.Run()
}

func sttyRestore(tty *os.File, saved string) {
	if saved == "" {
		return
	}
	cmd := exec.Command("stty", saved)
	cmd.Stdin = tty
	_ = cmd.Run()
}
