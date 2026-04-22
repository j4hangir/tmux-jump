package main

import (
	"fmt"
	"strings"
)

const (
	ansiReset     = "\x1b[0m"
	ansiDim       = "\x1b[2;37m"
	ansiMatch     = "\x1b[1;30;103m"
	ansiStatus    = "\x1b[7m"
	ansiClear     = "\x1b[2J\x1b[H"
	ansiHideCur   = "\x1b[?25l"
	ansiShowCur   = "\x1b[?25h"
)

func render(rows [][]rune, matches []Match, query string, width, height int) string {
	var b strings.Builder
	b.WriteString(ansiHideCur)
	b.WriteString(ansiClear)

	contentRows := height - 1
	if contentRows < 1 {
		contentRows = 1
	}

	hi := make([][]bool, len(rows))
	for _, m := range matches {
		if m.Row >= len(hi) {
			continue
		}
		if hi[m.Row] == nil {
			hi[m.Row] = make([]bool, len(rows[m.Row]))
		}
		for i := 0; i < m.Len && m.Col+i < len(hi[m.Row]); i++ {
			hi[m.Row][m.Col+i] = true
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
		var mask []bool
		if r < len(hi) {
			mask = hi[r]
		}
		writeRow(&b, row, mask, width)
	}

	b.WriteString("\r\n")
	b.WriteString(ansiStatus)
	status := fmt.Sprintf(" jump> %s  (%d matches) ", query, len(matches))
	if len(query) == 0 {
		status = fmt.Sprintf(" jump> (type to narrow; Esc to cancel) ")
	}
	if len(status) > width {
		status = status[:width]
	}
	b.WriteString(status)
	b.WriteString(strings.Repeat(" ", max(0, width-len(status))))
	b.WriteString(ansiReset)
	return b.String()
}

func writeRow(b *strings.Builder, row []rune, mask []bool, width int) {
	state := 0 // 0 dim, 1 match
	b.WriteString(ansiDim)
	limit := len(row)
	if limit > width {
		limit = width
	}
	for i := 0; i < limit; i++ {
		m := i < len(mask) && mask[i]
		if m && state == 0 {
			b.WriteString(ansiReset)
			b.WriteString(ansiMatch)
			state = 1
		} else if !m && state == 1 {
			b.WriteString(ansiReset)
			b.WriteString(ansiDim)
			state = 0
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
