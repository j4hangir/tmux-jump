package main

import "testing"

func rows(lines ...string) [][]rune {
	out := make([][]rune, len(lines))
	for i, l := range lines {
		out[i] = []rune(l)
	}
	return out
}

func TestFindMatchesEmpty(t *testing.T) {
	if got := findMatches(rows("hello"), []rune("")); got != nil {
		t.Fatalf("empty query: want nil, got %v", got)
	}
}

func TestFindMatchesSubstring(t *testing.T) {
	got := findMatches(rows("foo bar foo"), []rune("foo"))
	if len(got) != 2 {
		t.Fatalf("want 2 matches, got %d: %+v", len(got), got)
	}
	if got[0].Row != 0 || got[0].Col != 0 || got[0].Len != 3 {
		t.Errorf("match[0]=%+v", got[0])
	}
	if got[1].Row != 0 || got[1].Col != 8 {
		t.Errorf("match[1]=%+v", got[1])
	}
}

func TestFindMatchesSmartCaseLower(t *testing.T) {
	got := findMatches(rows("Foo foo FOO"), []rune("foo"))
	if len(got) != 3 {
		t.Fatalf("lowercase query should be case-insensitive; got %d matches: %+v", len(got), got)
	}
}

func TestFindMatchesSmartCaseMixed(t *testing.T) {
	got := findMatches(rows("Foo foo FOO"), []rune("Foo"))
	if len(got) != 1 {
		t.Fatalf("uppercase-containing query should be case-sensitive; got %d: %+v", len(got), got)
	}
	if got[0].Col != 0 {
		t.Errorf("want first Foo at col 0, got %+v", got[0])
	}
}

func TestFindMatchesAcrossRows(t *testing.T) {
	got := findMatches(rows("alpha", "beta", "gamma"), []rune("a"))
	// alpha: col 0, col 4 | beta: col 3 | gamma: col 1, col 4
	if len(got) != 5 {
		t.Fatalf("want 5 matches across rows, got %d: %+v", len(got), got)
	}
}

func TestFindMatchesOverlapping(t *testing.T) {
	// "aaaa" with query "aa" -> positions 0,1,2
	got := findMatches(rows("aaaa"), []rune("aa"))
	if len(got) != 3 {
		t.Fatalf("want 3 overlapping matches, got %d: %+v", len(got), got)
	}
}

func TestFindMatchesShortRow(t *testing.T) {
	got := findMatches(rows("hi"), []rune("hello"))
	if got != nil {
		t.Fatalf("row shorter than query should yield nothing, got %+v", got)
	}
}

func TestHasUpper(t *testing.T) {
	cases := map[string]bool{
		"foo":    false,
		"Foo":    true,
		"FOO":    true,
		"":       false,
		"123":    false,
		"café":   false,
		"CAFÉ":   true,
	}
	for in, want := range cases {
		if got := hasUpper([]rune(in)); got != want {
			t.Errorf("hasUpper(%q)=%v want %v", in, got, want)
		}
	}
}
