package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFontsHaveConsistentDimensions(t *testing.T) {
	for name, selected_font := range available_fonts {
		for r, glyph := range selected_font.glyphs {
			if len(glyph) != selected_font.height {
				t.Errorf("%s %q has %d rows, want %d", name, r, len(glyph), selected_font.height)
				continue
			}
			width := utf8.RuneCountInString(glyph[0])
			for row, line := range glyph[1:] {
				if got := utf8.RuneCountInString(line); got != width {
					t.Errorf("%s %q row %d has width %d, want %d", name, r, row+1, got, width)
				}
			}
		}
	}
}

func TestHyprblkFontMatchesSample(t *testing.T) {
	lines := get_text("[(1+2)/3]=1*0", "", "", available_fonts["hyprblk"])
	want := []string{
		"▄▄▄   ▄                ▄     ▄      ▄▄▄",
		"█   ▄▀  ▄█    █   ▀▀▀█  ▀▄  ▄▀ ▀▀▀█   █ ▄▄▄ ▄█  ▄▀▄▀▄ █▀▀█",
		"█   █    █  ▀▀█▀▀ █▀▀▀   █  █  ▀▀▀█   █ ▄▄▄  █  ▄▀█▀▄ █  █",
		"█▄▄  ▀▄ ▀▀▀   ▀   ▀▀▀▀ ▄▀  █   ▀▀▀▀ ▄▄█     ▀▀▀  ▀ ▀  ▀▀▀▀",
	}
	for i := range want {
		if got := strings.TrimRight(lines[i], " "); got != want[i] {
			t.Errorf("line %d:\n got %q\nwant %q", i, got, want[i])
		}
	}
}

func TestHyprtxtFontMatchesSample(t *testing.T) {
	lines := get_text("[(1+2)/3]=1*0", "", "", available_fonts["hyprtxt"])
	want := []string{
		"█▀ ▄█ ░▄░ ▀█  █ ▀▀█ ▀█ ▄▄ ▄█ ▄░▄ █▀█",
		"█▄  █ ▀█▀ █▄ █  ▄██ ▄█ ▄▄  █ ▄▀▄ █▄█",
	}
	for i := range want {
		if got := strings.TrimRight(lines[i], " "); got != want[i] {
			t.Errorf("line %d:\n got %q\nwant %q", i, got, want[i])
		}
	}
}

func TestHyprblkFontSupportsHyprtxtCharset(t *testing.T) {
	for r := range font_hprtxt {
		if _, ok := font_hyprblk[r]; !ok {
			t.Errorf("hyprblk font is missing %q", r)
		}
	}
}

func TestParseFontOption(t *testing.T) {
	options, args, err := parse_options([]string{"--font", "hyprblk", "opencode"})
	if err != nil {
		t.Fatal(err)
	}
	if options.font_name != "hyprblk" {
		t.Fatalf("got font %q, want hyprblk", options.font_name)
	}
	if len(args) != 1 || args[0] != "opencode" {
		t.Fatalf("unexpected arguments: %q", args)
	}
}

func TestDefaultFontName(t *testing.T) {
	options, _, err := parse_options(nil)
	if err != nil {
		t.Fatal(err)
	}
	if options.font_name != "hyprtxt" {
		t.Fatalf("got default font %q, want hyprtxt", options.font_name)
	}
}
