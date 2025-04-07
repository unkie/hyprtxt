# hyprtxt

A minimalistic CLI tool to render text to the console using a custom 'hyprfont' figlet font.

---

## Features

- Render any text in a 2-line ASCII font.
- Output the font as a figlet `.flf` file.
- Check if any characters in the input are missing from the font.
- Show a gallery of the characters supported by the font.

---

## Usage

```sh
hyprtxt [options] [text]
```

### Options

| Flag        | Description                                    |
|-------------|------------------------------------------------|
| `-figlet`   | Output the embedded font in figlet format      |
| `-missing`  | Show unsupported characters in the input       |
| `-charset`  | Print supported character set                  |
| `-examples` | Print all characters as a glyph gallery        |
| `-version`  | Show version info                              |
| `-help`     | Show help message                              |

---

## Examples

```sh
hyprtxt hello world
hyprtxt -missing "{ oh no !¡}"
hyprtxt -figlet > hyprfont.flf
```

---

## Development

### Build

```sh
make build
```

### Run

```sh
make run ARGS="hello world"
```

---

## License

MIT © 2025 [Mark Pustjens](mailto:pustjens@dds.nl)

