package main

import (
	"flag"
	"fmt"
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
	flag.Parse()
	if *capturePath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: tmux-jump -capture FILE -out FILE [-w W] [-h H]")
		os.Exit(2)
	}

	data, err := os.ReadFile(*capturePath)
	if err != nil {
		die(err)
	}
	rows := parseRows(string(data))

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		die(err)
	}
	defer tty.Close()

	saved, err := sttySave(tty)
	if err != nil {
		die(err)
	}
	defer sttyRestore(tty, saved)
	if err := sttyRaw(tty); err != nil {
		die(err)
	}
	defer fmt.Fprint(tty, ansiShowCur+ansiReset)

	query := []rune{}
	var matches []Match

	for {
		fmt.Fprint(tty, render(rows, matches, string(query), *width, *height))

		buf := make([]byte, 1)
		n, err := tty.Read(buf)
		if err != nil || n == 0 {
			return
		}
		b := buf[0]

		switch {
		case b == 0x1b, b == 0x03, b == 0x07:
			return
		case b == 0x7f, b == 0x08:
			if len(query) > 0 {
				query = query[:len(query)-1]
				matches = findMatches(rows, query)
			}
		case b == '\r', b == '\n':
			if len(matches) >= 1 {
				writeResult(matches[0])
				return
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
			if len(matches) == 1 {
				writeResult(matches[0])
				return
			}
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
	cmd := exec.Command("stty", "-icanon", "-echo", "min", "1", "time", "0")
	cmd.Stdin = tty
	return cmd.Run()
}

func sttyRestore(tty *os.File, saved string) {
	cmd := exec.Command("stty", saved)
	cmd.Stdin = tty
	_ = cmd.Run()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
