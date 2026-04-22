package main

import "unicode"

type Match struct {
	Row, Col, Len int
}

func findMatches(rows [][]rune, query []rune) []Match {
	if len(query) == 0 {
		return nil
	}
	cs := hasUpper(query)
	var out []Match
	for r, row := range rows {
		if len(row) < len(query) {
			continue
		}
		for c := 0; c+len(query) <= len(row); c++ {
			if equalAt(row, c, query, cs) {
				out = append(out, Match{Row: r, Col: c, Len: len(query)})
			}
		}
	}
	return out
}

func hasUpper(s []rune) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func equalAt(row []rune, c int, query []rune, cs bool) bool {
	for i, q := range query {
		a := row[c+i]
		if cs {
			if a != q {
				return false
			}
		} else {
			if unicode.ToLower(a) != unicode.ToLower(q) {
				return false
			}
		}
	}
	return true
}
