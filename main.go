// Copyright (c) 2025 Mark Pustjens <pustjens@dds.nl>

package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/pflag"
)

var version = "0.1.0"

func get_text(input string, prefix string, postfix string) []string {
	lines := []string{prefix, prefix}
	first := true
	for _, r := range input {
		glyph, ok := font[r]
		if !ok {
			continue
		}
		if !first {
			lines[0] += " "
			lines[1] += " "
		}
		lines[0] += glyph[0]
		lines[1] += glyph[1]
		first = false
	}
	lines[0] += postfix
	lines[1] += postfix

	return lines
}

func render(input string, prefix string, postfix string) {
	lines := get_text(input, prefix, postfix)
	fmt.Println(lines[0])
	fmt.Println(lines[1])
}

func check_missing(input string) {
	var missing []string
	for _, r := range input {
		if _, ok := font[r]; !ok {
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

func print_charset() {
	keys := make([]rune, 0, len(font))
	for r := range font {
		keys = append(keys, r)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, r := range keys {
		fmt.Printf("%c ", r)
	}
	fmt.Println()
}

func print_examples() {
	keys := sorted_keys(font)
	for i := 0; i < len(keys); i += 16 {
		end := i + 16
		if end > len(keys) {
			end = len(keys)
		}
		label := ""
		line1 := ""
		line2 := ""
		for j, r := range keys[i:end] {
			glyph := font[r]
			if r == ' ' {
				r = '⎵'
			}
			label += fmt.Sprintf("%c%s", r, strings.Repeat(" ", utf8.RuneCountInString(glyph[0])-1))
			line1 += glyph[0]
			line2 += glyph[1]
			if j != 15 {
				line1 += "  "
				line2 += "  "
				label += "  "
			}
		}
		fmt.Println(label)
		fmt.Println(line1)
		fmt.Println(line2)
		fmt.Println()
	}
}

func print_flf() {
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

	fmt.Println("flf2a$ 2 2 8 0 3 0 64 0")
	fmt.Println("Font Author: Mark Pustjens")
	fmt.Println("")
	fmt.Println("FIGFont created with: https://github.com/unkie/hyprtxt")
	for _, i := range charset {
		r := unicode.ToLower(rune(i))
		g, ok := font[r]
		if !ok {
			g = []string{"", ""}
		}
		if r == ' ' {
			g = []string{"$", "$"}
		}
		fmt.Printf("%s$@\n%s$@@\n", g[0], g[1])
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

When used without options, outputs the text with the 2-line
hyprfont font. All input is converted to lowercase.
Unsupported characters are omitted in the output.

Options:
	-p, --prefix <text>
		Prefix each output line.
	-P, --postfix <text>
		Postfix each output line.
	-f, --figlet
		Output font in figlet .flf format
	-m, --missing
		Show unsupported characters in input
	-c, --charset
		Print the supported character in ASCII
	-e, --examples
		Print the supported characters in hyprfont
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

	switch {
	case options.show_flf:
		print_flf()
	case options.show_missing:
		check_missing(text)
	case options.show_charset:
		print_charset()
	case options.show_examples:
		print_examples()
	case options.show_version:
		fmt.Println("hyprtxt version", version)
	case options.show_help, text == "":
		print_help()
	default:
		render(text, options.prefix, options.postfix)
	}
}
