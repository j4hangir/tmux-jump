package main

import (
	"fmt"
	"strings"
)

const (
	ansiReset    = "\x1b[0m"
	ansiDim      = "\x1b[2;37m"
	ansiMatch    = "\x1b[1;30;103m"
	ansiSelected = "\x1b[1;97;41m"
	ansiStatus   = "\x1b[7m"
	ansiClear    = "\x1b[2J\x1b[H"
	ansiHideCur  = "\x1b[?25l"
	ansiShowCur  = "\x1b[?25h"
)

// cell: 0 none, 1 match, 2 selected match
func render(rows [][]rune, matches []Match, selected int, query string, width, height int) string {
	var b strings.Builder
	b.WriteString(ansiHideCur)
	b.WriteString(ansiClear)

	contentRows := height - 1
	if contentRows < 1 {
		contentRows = 1
	}

	navigable := len(matches) > 1 && len(matches) <= 9

	hi := make([][]byte, len(rows))
	for i, m := range matches {
		if m.Row >= len(hi) {
			continue
		}
		if hi[m.Row] == nil {
			hi[m.Row] = make([]byte, len(rows[m.Row]))
		}
		mark := byte(1)
		if navigable && i == selected {
			mark = 2
		}
		for k := 0; k < m.Len && m.Col+k < len(hi[m.Row]); k++ {
			hi[m.Row][m.Col+k] = mark
		}
	}

	for r := 0; r < contentRows; r++ {
		if r > 0 {
			b.WriteString("\r\n")
		}
		if r >= len(rows) {
			continue
		}
		row := rows[r]
		var mask []byte
		if r < len(hi) {
			mask = hi[r]
		}
		writeRow(&b, row, mask, width)
	}

	b.WriteString("\r\n")
	b.WriteString(ansiStatus)
	var status string
	switch {
	case len(query) == 0:
		status = " jump> (type to narrow; Esc to cancel) "
	case navigable:
		status = fmt.Sprintf(" jump> %s  [%d/%d]  Tab/↑↓ to pick, Enter to jump ", query, selected+1, len(matches))
	default:
		status = fmt.Sprintf(" jump> %s  (%d matches) ", query, len(matches))
	}
	if len(status) > width {
		status = status[:width]
	}
	b.WriteString(status)
	b.WriteString(strings.Repeat(" ", max(0, width-len(status))))
	b.WriteString(ansiReset)
	return b.String()
}

func writeRow(b *strings.Builder, row []rune, mask []byte, width int) {
	state := byte(0)
	b.WriteString(ansiDim)
	limit := len(row)
	if limit > width {
		limit = width
	}
	for i := 0; i < limit; i++ {
		var m byte
		if i < len(mask) {
			m = mask[i]
		}
		if m != state {
			b.WriteString(ansiReset)
			switch m {
			case 0:
				b.WriteString(ansiDim)
			case 1:
				b.WriteString(ansiMatch)
			case 2:
				b.WriteString(ansiSelected)
			}
			state = m
		}
		b.WriteRune(row[i])
	}
	b.WriteString(ansiReset)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
