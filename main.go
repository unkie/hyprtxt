// Copyright (c) 2025 Mark Pustjens <pustjens@dds.nl>

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/pflag"
)

var version = "devel"

type font_definition struct {
	height int
	glyphs map[rune][]string
}

var available_fonts = map[string]font_definition{
	"hyprtxt": {height: 2, glyphs: font_hprtxt},
	"hyprblk": {height: 4, glyphs: font_hyprblk},
}

func get_build_version() string {
	if version != "devel" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func get_text(input string, prefix string, postfix string, selected_font font_definition) []string {
	lines := make([]string, selected_font.height)
	for i := range lines {
		lines[i] = prefix
	}
	first := true
	for _, r := range input {
		glyph, ok := selected_font.glyphs[r]
		if !ok {
			continue
		}
		if !first {
			for i := range lines {
				lines[i] += " "
			}
		}
		for i := range lines {
			lines[i] += glyph[i]
		}
		first = false
	}
	for i := range lines {
		lines[i] += postfix
	}

	return lines
}

func render(input string, prefix string, postfix string, selected_font font_definition) {
	for _, line := range get_text(input, prefix, postfix, selected_font) {
		fmt.Println(line)
	}
}

func check_missing(input string, selected_font font_definition) {
	var missing []string
	for _, r := range input {
		if _, ok := selected_font.glyphs[r]; !ok {
			missing = append(missing, fmt.Sprintf("%c", r))
		}
	}
	if len(missing) == 0 {
		fmt.Println("All characters supported.")
		return
	}
	fmt.Println("Unsupported characters:")
	fmt.Println(strings.Join(missing, ", "))
}

func print_charset(selected_font font_definition) {
	keys := make([]rune, 0, len(selected_font.glyphs))
	for r := range selected_font.glyphs {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, r := range keys {
		fmt.Printf("%c ", r)
	}
	fmt.Println()
}

func print_examples(selected_font font_definition) {
	keys := sorted_keys(selected_font.glyphs)
	for i := 0; i < len(keys); i += 16 {
		end := i + 16
		if end > len(keys) {
			end = len(keys)
		}
		label := ""
		lines := make([]string, selected_font.height)
		for j, r := range keys[i:end] {
			glyph := selected_font.glyphs[r]
			if r == ' ' {
				r = '⎵'
			}
			label += fmt.Sprintf("%c%s", r, strings.Repeat(" ", utf8.RuneCountInString(glyph[0])-1))
			for row := range lines {
				lines[row] += glyph[row]
			}
			if j != end-i-1 {
				for row := range lines {
					lines[row] += "  "
				}
				label += "  "
			}
		}
		fmt.Println(label)
		for _, line := range lines {
			fmt.Println(line)
		}
		fmt.Println()
	}
}

func print_flf(selected_font font_definition) {
	charset := []rune{
		' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-',
		'.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';',
		'<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
		'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
		'X', 'Y', 'Z', '[', '\\', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e',
		'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's',
		't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', 'Ä', 'Ö', 'Ü',
		'ä', 'ö', 'ü', 'ß',
	}

	max_length := 0
	for _, glyph := range selected_font.glyphs {
		for _, line := range glyph {
			length := utf8.RuneCountInString(line) + 3
			if length > max_length {
				max_length = length
			}
		}
	}
	fmt.Printf("flf2a$ %d %d %d 0 3 0 64 0\n", selected_font.height, selected_font.height, max_length)
	fmt.Println("Font Author: Mark Pustjens")
	fmt.Println("")
	fmt.Println("FIGFont created with: https://github.com/unkie/hyprtxt")
	for _, i := range charset {
		r := unicode.ToLower(rune(i))
		g, ok := selected_font.glyphs[r]
		if !ok {
			g = make([]string, selected_font.height)
		}
		if r == ' ' {
			g = make([]string, selected_font.height)
			for row := range g {
				g[row] = "$"
			}
		}
		for row, line := range g {
			endmark := "@"
			if row == len(g)-1 {
				endmark = "@@"
			}
			fmt.Printf("%s$%s\n", line, endmark)
		}
	}
}

func sorted_keys(m map[rune][]string) []rune {
	var keys []rune
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

type cli_options struct {
	prefix        string
	postfix       string
	font_name     string
	show_flf      bool
	show_missing  bool
	show_charset  bool
	show_examples bool
	show_version  bool
	show_help     bool
}

func parse_options(args []string) (cli_options, []string, error) {
	var options cli_options
	flags := pflag.NewFlagSet("hyprtxt", pflag.ContinueOnError)
	flags.SetInterspersed(false)
	flags.StringVarP(&options.prefix, "prefix", "p", "", "")
	flags.StringVarP(&options.postfix, "postfix", "P", "", "")
	flags.StringVarP(&options.font_name, "font", "F", "hyprtxt", "")
	flags.BoolVarP(&options.show_flf, "figlet", "f", false, "")
	flags.BoolVarP(&options.show_missing, "missing", "m", false, "")
	flags.BoolVarP(&options.show_charset, "charset", "c", false, "")
	flags.BoolVarP(&options.show_examples, "examples", "e", false, "")
	flags.BoolVarP(&options.show_version, "version", "v", false, "")
	flags.BoolVarP(&options.show_help, "help", "h", false, "")

	if err := flags.Parse(args); err != nil {
		return cli_options{}, nil, err
	}
	return options, flags.Args(), nil
}

func print_help() {
	fmt.Println(`Usage: hyprtxt [options] [text]

When used without an action option, renders text with the selected
font. The default font is hyprtxt. All input is converted to lowercase.
Unsupported characters are omitted in the output.

Options:
	-p, --prefix <text>
		Prefix each output line.
	-P, --postfix <text>
		Postfix each output line.
	-F, --font <name>
		Select font: hyprtxt or hyprblk.
	-f, --figlet
		Output font in figlet .flf format
	-m, --missing
		Show unsupported characters in input
	-c, --charset
		Print the supported character in ASCII
	-e, --examples
		Print the selected font's supported characters
	-v, --version
		Show version info
	-h, --help
		Show this help`)
}

func main() {
	options, args, err := parse_options(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "hyprtxt:", err)
		os.Exit(2)
	}
	text := strings.ToLower(strings.Join(args, " "))
	selected_font, ok := available_fonts[options.font_name]
	if !ok {
		fmt.Fprintf(os.Stderr, "hyprtxt: unknown font %q (available: hyprtxt, hyprblk)\n", options.font_name)
		os.Exit(2)
	}

	switch {
	case options.show_flf:
		print_flf(selected_font)
	case options.show_missing:
		check_missing(text, selected_font)
	case options.show_charset:
		print_charset(selected_font)
	case options.show_examples:
		print_examples(selected_font)
	case options.show_version:
		fmt.Println("hyprtxt version", get_build_version())
	case options.show_help, text == "":
		print_help()
	default:
		render(text, options.prefix, options.postfix, selected_font)
	}
}
