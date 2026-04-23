package main

import (
	"strings"
	"testing"
)

func TestHintIndex(t *testing.T) {
	hints := []rune("duhetonasi")
	cases := map[rune]int{
		'd': 0,
		'u': 1,
		'i': 9,
		'x': -1,
	}
	for r, want := range cases {
		if got := hintIndex(hints, r); got != want {
			t.Errorf("hintIndex(%q)=%d want %d", r, got, want)
		}
	}
}

func TestDefaultSelectedWithLimit10(t *testing.T) {
	cases := map[int]int{
		0:  0,
		1:  0,
		10: 9,
		11: 0, // beyond limit → 0
	}
	for n, want := range cases {
		if got := defaultSelected(n); got != want {
			t.Errorf("defaultSelected(%d)=%d want %d", n, got, want)
		}
	}
}

func TestRenderHintBadgesOverlayFirstChar(t *testing.T) {
	rs := rows("foo bar foo baz foo")
	ms := findMatches(rs, []rune("foo"))
	if len(ms) != 3 {
		t.Fatalf("setup: want 3 matches, got %d", len(ms))
	}
	hints := []rune("duhetonasi")
	out := render(rs, ms, 0, "foo", 80, 24, true, hints)
	// First match gets 'd', second 'u', third 'h'
	if !strings.Contains(out, "d") || !strings.Contains(out, "u") || !strings.Contains(out, "h") {
		t.Fatalf("hint glyphs missing from render output:\n%s", out)
	}
	// Status bar should mention hint mode
	if !strings.Contains(out, "hint") {
		t.Errorf("status bar should mention hint mode:\n%s", out)
	}
}

func TestRenderNoHintsWhenDisabled(t *testing.T) {
	rs := rows("foo bar foo")
	ms := findMatches(rs, []rune("foo"))
	out := render(rs, ms, 0, "foo", 80, 24, false, []rune("duhetonasi"))
	// Status bar should use navigable format, not hint mode
	if strings.Contains(out, "hint: press") {
		t.Errorf("hint mode status leaked into non-hint render")
	}
}

func TestRenderNavigableAtTenMatches(t *testing.T) {
	// 10 matches: "a" in 10 distinct rows
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "a"
	}
	rs := rows(lines...)
	ms := findMatches(rs, []rune("a"))
	if len(ms) != 10 {
		t.Fatalf("setup: want 10 matches, got %d", len(ms))
	}
	out := render(rs, ms, 0, "a", 80, 24, false, []rune("duhetonasi"))
	if !strings.Contains(out, "[1/10]") {
		t.Errorf("expected navigable status for 10 matches, got:\n%s", out)
	}
}
