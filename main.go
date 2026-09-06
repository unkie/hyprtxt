// Copyright (c) 2025 Mark Pustjens <pustjens@dds.nl>

package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/pflag"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
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

const reset_background = "\x1b[49m"

type rgb_color struct {
	red   uint8
	green uint8
	blue  uint8
}

func (color rgb_color) darker() rgb_color {
	return rgb_color{
		red:   uint8(int(color.red) * 35 / 100),
		green: uint8(int(color.green) * 35 / 100),
		blue:  uint8(int(color.blue) * 35 / 100),
	}
}

func (color rgb_color) background_sequence() string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", color.red, color.green, color.blue)
}

func (color rgb_color) foreground_sequence() string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", color.red, color.green, color.blue)
}

func parse_terminal_color(response string) (rgb_color, bool) {
	start := strings.Index(response, "rgb:")
	if start == -1 {
		return rgb_color{}, false
	}
	value := response[start+len("rgb:"):]
	if end := strings.IndexAny(value, "\a\x1b"); end != -1 {
		value = value[:end]
	}
	components := strings.Split(value, "/")
	if len(components) != 3 {
		return rgb_color{}, false
	}

	values := [3]uint8{}
	for i, component := range components {
		if len(component) < 1 || len(component) > 4 {
			return rgb_color{}, false
		}
		parsed, err := strconv.ParseUint(component, 16, 16)
		if err != nil {
			return rgb_color{}, false
		}
		maximum := uint64(1)<<(4*len(component)) - 1
		values[i] = uint8(parsed * 255 / maximum)
	}
	return rgb_color{red: values[0], green: values[1], blue: values[2]}, true
}

func parse_sgr_foreground(response string) (rgb_color, bool, bool) {
	start := strings.Index(response, "$r")
	if start == -1 {
		return rgb_color{}, false, false
	}
	value := response[start+len("$r"):]
	if end := strings.Index(value, "m"); end != -1 {
		value = value[:end]
	}
	parameters := strings.Split(value, ";")
	for i, parameter := range parameters {
		if strings.HasPrefix(parameter, "38:2:") {
			parts := strings.Split(parameter, ":")
			if len(parts) < 5 {
				return rgb_color{}, false, true
			}
			color, ok := parse_rgb_components(parts[len(parts)-3:])
			return color, ok, true
		}
		if parameter == "38" {
			if i+1 >= len(parameters) || parameters[i+1] != "2" || i+4 >= len(parameters) {
				return rgb_color{}, false, true
			}
			color, ok := parse_rgb_components(parameters[i+2 : i+5])
			return color, ok, true
		}
		code, err := strconv.Atoi(parameter)
		if strings.HasPrefix(parameter, "38:") || err == nil && (code >= 30 && code <= 37 || code >= 90 && code <= 97) {
			return rgb_color{}, false, true
		}
	}
	return rgb_color{}, false, false
}

func parse_rgb_components(components []string) (rgb_color, bool) {
	values := [3]uint8{}
	for i, component := range components {
		parsed, err := strconv.ParseUint(component, 10, 8)
		if err != nil {
			return rgb_color{}, false
		}
		values[i] = uint8(parsed)
	}
	return rgb_color{red: values[0], green: values[1], blue: values[2]}, true
}

func terminal_query(request string) (string, bool) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return "", false
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", false
	}
	defer tty.Close()

	state, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return "", false
	}
	defer term.Restore(int(tty.Fd()), state)

	if _, err := tty.WriteString(request); err != nil {
		return "", false
	}

	deadline := time.Now().Add(150 * time.Millisecond)
	response := make([]byte, 0, 64)
	for time.Now().Before(deadline) {
		wait := max(int(time.Until(deadline).Milliseconds()), 1)
		fds := []unix.PollFd{{Fd: int32(tty.Fd()), Events: unix.POLLIN}}
		ready, err := unix.Poll(fds, wait)
		if err != nil || ready == 0 {
			break
		}
		buffer := make([]byte, 64)
		n, err := tty.Read(buffer)
		if err != nil {
			break
		}
		response = append(response, buffer[:n]...)
		if strings.Contains(string(response), "\a") || strings.Contains(string(response), "\x1b\\") {
			break
		}
	}
	return string(response), len(response) > 0
}

func terminal_foreground() (rgb_color, bool) {
	if response, ok := terminal_query("\x1bP$qm\x1b\\"); ok {
		color, parsed, active := parse_sgr_foreground(response)
		if parsed {
			return color, true
		}
		if active {
			return rgb_color{}, false
		}
	}
	response, ok := terminal_query("\x1b]10;?\x07")
	if !ok {
		return rgb_color{}, false
	}
	return parse_terminal_color(response)
}

func add_background(glyphs map[rune][]string, foreground *rgb_color) map[rune][]string {
	if foreground == nil {
		return glyphs
	}

	const (
		normal = iota
		half_background
		cell_background
	)

	background := foreground.darker()
	background_foreground_sequence := background.foreground_sequence()
	background_sequence := background.background_sequence()
	restore_foreground_sequence := foreground.foreground_sequence()

	with_background := make(map[rune][]string, len(glyphs))
	for r, glyph := range glyphs {
		lines := append([]string(nil), glyph...)
		has_foreground_above := make([]bool, utf8.RuneCountInString(glyph[0]))
		for row := range lines {
			current := []rune(glyph[row])
			var line strings.Builder
			mode := normal
			for column, cell := range current {
				next_mode := normal
				output := cell

				top_foreground := cell == '█' || cell == '▀'
				bottom_foreground := cell == '█' || cell == '▄'
				top_pixel := row * 2
				bottom_pixel := top_pixel + 1

				// Only pixel rows 5-7 can have the darker background, and only
				// after a foreground pixel has appeared above it in this column.
				background_top := top_pixel >= 4 && top_pixel <= 6 && has_foreground_above[column]
				has_foreground_above[column] = has_foreground_above[column] || top_foreground
				background_bottom := bottom_pixel >= 4 && bottom_pixel <= 6 && has_foreground_above[column]
				has_foreground_above[column] = has_foreground_above[column] || bottom_foreground

				background_top = background_top && !top_foreground
				background_bottom = background_bottom && !bottom_foreground
				switch {
				case background_top && background_bottom:
					next_mode = cell_background
					output = ' '
				case background_top && cell == '▄', background_bottom && cell == '▀':
					next_mode = cell_background
				case background_top:
					next_mode = half_background
					output = '▀'
				case background_bottom:
					next_mode = half_background
					output = '▄'
				}
				// Change ANSI state only when the rendering mode changes. Besides keeping
				// the output compact, the targeted resets preserve the caller's active
				// foreground color between background and glyph runs.
				if next_mode != mode {
					switch mode {
					case half_background:
						line.WriteString(restore_foreground_sequence)
					case cell_background:
						line.WriteString(reset_background)
					}
					switch next_mode {
					case half_background:
						line.WriteString(background_foreground_sequence)
					case cell_background:
						line.WriteString(background_sequence)
					}
					mode = next_mode
				}
				line.WriteRune(output)
			}
			switch mode {
			case half_background:
				line.WriteString(restore_foreground_sequence)
			case cell_background:
				line.WriteString(reset_background)
			}
			lines[row] = line.String()
		}
		with_background[r] = lines
	}
	return with_background
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

func render(input string, prefix string, postfix string, selected_font font_definition, background bool) {
	if !background {
		for _, line := range get_text(input, prefix, postfix, selected_font) {
			fmt.Println(line)
		}
		return
	}

	fmt.Print(prefix)
	foreground, ok := terminal_foreground()
	var foreground_color *rgb_color
	if ok {
		foreground_color = &foreground
	}
	selected_font.glyphs = add_background(selected_font.glyphs, foreground_color)
	for row, line := range get_text(input, "", postfix, selected_font) {
		if row > 0 {
			fmt.Print(prefix)
		}
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

func print_examples(selected_font font_definition, background bool) {
	keys := sorted_keys(selected_font.glyphs)
	rendered_glyphs := selected_font.glyphs
	if background {
		foreground, ok := terminal_foreground()
		var foreground_color *rgb_color
		if ok {
			foreground_color = &foreground
		}
		rendered_glyphs = add_background(selected_font.glyphs, foreground_color)
	}
	for i := 0; i < len(keys); i += 16 {
		end := i + 16
		if end > len(keys) {
			end = len(keys)
		}
		label := ""
		lines := make([]string, selected_font.height)
		for j, r := range keys[i:end] {
			glyph := selected_font.glyphs[r]
			rendered_glyph := rendered_glyphs[r]
			if r == ' ' {
				r = '⎵'
			}
			label += fmt.Sprintf("%c%s", r, strings.Repeat(" ", utf8.RuneCountInString(glyph[0])-1))
			for row := range lines {
				lines[row] += rendered_glyph[row]
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
	background    bool
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
	flags.BoolVarP(&options.background, "background", "b", false, "")
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
	-b, --background
		Add a darker background to the hyprblk font.
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
	if options.background && options.font_name != "hyprblk" {
		fmt.Fprintln(os.Stderr, "hyprtxt: --background is only available with --font hyprblk")
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
		print_examples(selected_font, options.background)
	case options.show_version:
		fmt.Println("hyprtxt version", get_build_version())
	case options.show_help, text == "":
		print_help()
	default:
		render(text, options.prefix, options.postfix, selected_font, options.background)
	}
}
