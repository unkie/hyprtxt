// Copyright (c) 2025 Mark Pustjens <pustjens@dds.nl>

package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var version = "0.1.0"

func render(input string) {
	lines := []string{"", ""}
	input = strings.ToLower(input)
	for i, r := range input {
		glyph, ok := font[r]
		if !ok {
			continue
		}
		lines[0] += glyph[0]
		lines[1] += glyph[1]
		if i != len(input)-1 {
			lines[0] += " "
			lines[1] += " "
		}
	}
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
	var charset = []rune {
		' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-',
		'.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';',
		'<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
		'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
		'X', 'Y', 'Z', '[', '\\', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e',
		'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's',
		't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', 'Ä', 'Ö', 'Ü',
		'ä', 'ö', 'ü', 'ß',
	}

	fmt.Println("flf2a$ 4 3 7 0 3 0 64 0")
	fmt.Println("Font Author: Mark Pustjens")
	fmt.Println("")
	fmt.Println("FIGFont created with: https://github.com/unkie/hyprtxt")
	for _, i := range charset {
		var r = unicode.ToLower (rune(i))
		g, ok := font[r]
		if !ok {
			g = []string{"", ""}
		}
		if r == ' ' {
			g = []string{"\u2003", "\u2003"}
		}
		fmt.Printf("%s@\n%s@\n%s@\n%s@@\n",
		g[0], g[1],
		strings.Repeat(" ", utf8.RuneCountInString(g[0])),
		strings.Repeat(" ", utf8.RuneCountInString(g[0])))
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

func print_help() {
	fmt.Println(`Usage: hyprtext [options] [text]

When used without options, outputs the text with the 2-line
hyprfont font. All input is converted to lowercase.
Unsupported characters are omitted in the output.

Options:
    -figlet
        Output font in figlet .flf format
    -missing
        Show unsupported characters in input
    -charset
        Print the supported character in ASCII
    -examples
        Print the supported characters in hyprfont
    -version
        Show version info
    -help
        Show this help`)
}

func main() {
	show_flf := flag.Bool("figlet", false, "")
	show_missing := flag.Bool("missing", false, "")
	show_charset := flag.Bool("charset", false, "")
	show_examples := flag.Bool("examples", false, "")
	show_version := flag.Bool("version", false, "")
	show_help := flag.Bool("help", false, "")
	flag.Usage = print_help
	flag.Parse()

	args := flag.Args()
	text := strings.Join(args, "")

	switch {
	case *show_flf:
		print_flf()
	case *show_missing:
		check_missing(text)
	case *show_charset:
		print_charset()
	case *show_examples:
		print_examples()
	case *show_version:
		fmt.Println("hyprtxt version", version)
	case *show_help, text == "":
		print_help()
	default:
		render(text)
	}
}
