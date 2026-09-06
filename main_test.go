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

func TestHyprblkFontAddsANSIBackground(t *testing.T) {
	color := rgb_color{red: 100, green: 150, blue: 200}
	with_background := add_background(map[rune][]string{
		'o': {"    ", "█▀▀█", "█  █", "▀▀▀▀"},
		' ': {" ", " ", " ", " "},
		'_': {"    ", "    ", "    ", "▀▀▀▀"},
	}, &color)
	background := color.darker()
	want_top := "█▀▀█"
	if got := with_background['o'][1]; got != want_top {
		t.Errorf("top: got %q, want %q", got, want_top)
	}
	want_middle := "█" + background.background_sequence() + "  " + reset_background + "█"
	if got := with_background['o'][2]; got != want_middle {
		t.Errorf("middle: got %q, want %q", got, want_middle)
	}
	if got, want := with_background[' '][2], " "; got != want {
		t.Errorf("space: got %q, want %q", got, want)
	}
	if got, want := with_background['_'][2], "    "; got != want {
		t.Errorf("underscore middle: got %q, want %q", got, want)
	}
	if got, want := with_background['_'][3], "▀▀▀▀"; got != want {
		t.Errorf("underscore bottom: got %q, want %q", got, want)
	}
}

func TestHyprblkBackgroundHandlesHalfBlocks(t *testing.T) {
	color := rgb_color{red: 100, green: 150, blue: 200}
	with_background := add_background(map[rune][]string{
		'e': {"    ", "█▀▀█", "█▀▀▀", "▀▀▀▀"},
		'n': {"    ", "█▀▀▄", "█  █", "▀  ▀"},
	}, &color)
	background := color.darker()
	want_e := "█" + background.background_sequence() + "▀▀▀" + reset_background
	if got := with_background['e'][2]; got != want_e {
		t.Errorf("e: got %q, want %q", got, want_e)
	}
	want_n_middle := "█" + background.background_sequence() + "  " + reset_background + "█"
	if got := with_background['n'][2]; got != want_n_middle {
		t.Errorf("n middle: got %q, want %q", got, want_n_middle)
	}
	want_n_bottom := "▀" + background.foreground_sequence() + "▀▀" + color.foreground_sequence() + "▀"
	if got := with_background['n'][3]; got != want_n_bottom {
		t.Errorf("n bottom: got %q, want %q", got, want_n_bottom)
	}
	if got, want := with_background['n'][1], "█▀▀▄"; got != want {
		t.Errorf("n top: got %q, want %q", got, want)
	}
}

func TestHyprblkBackgroundHandlesEveryCell(t *testing.T) {
	color := rgb_color{red: 100, green: 150, blue: 200}
	with_background := add_background(map[rune][]string{
		'x': {"    ", "████", "█▀▄ ", "    "},
	}, &color)
	background := color.darker()
	want := "█" + background.background_sequence() + "▀▄ " + reset_background
	if got := with_background['x'][2]; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseTerminalColor(t *testing.T) {
	color, ok := parse_terminal_color("\x1b]10;rgb:8080/4000/ffff\x1b\\")
	if !ok {
		t.Fatal("terminal color was not parsed")
	}
	want := rgb_color{red: 128, green: 63, blue: 255}
	if color != want {
		t.Errorf("got %+v, want %+v", color, want)
	}
}

func TestParseSGRForeground(t *testing.T) {
	tests := []string{
		"\x1bP1$r0;38;2;120;180;255m\x1b\\",
		"\x1bP1$r0;38:2::120:180:255m\x1b\\",
	}
	for _, response := range tests {
		color, ok, active := parse_sgr_foreground(response)
		if !ok || !active {
			t.Fatalf("active terminal color was not parsed from %q", response)
		}
		want := rgb_color{red: 120, green: 180, blue: 255}
		if color != want {
			t.Errorf("got %+v, want %+v", color, want)
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
	options, args, err := parse_options([]string{"--font", "hyprblk", "-b", "opencode"})
	if err != nil {
		t.Fatal(err)
	}
	if options.font_name != "hyprblk" {
		t.Fatalf("got font %q, want hyprblk", options.font_name)
	}
	if !options.background {
		t.Fatal("background option was not set")
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
	if options.background {
		t.Fatal("background should be disabled by default")
	}
}
