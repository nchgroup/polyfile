# polyfile

A command-line tool for creating **polyglot files** - single binary artifacts simultaneously
valid in two or more file formats.

```
polyfile build document.pdf archive.zip -o combo.pdf
```

`combo.pdf` opens as a PDF in any viewer. `unzip combo.pdf` extracts the ZIP contents.

---

## How it works

Most file parsers are liberal: they look for their own magic bytes or structural markers and
ignore everything else. polyfile performs the byte surgery required to satisfy two parsers at
once — patching internal offsets, injecting guard bytes, or encapsulating payloads where needed.

---

## Supported format pairs

| Pair      | Strategy       | Notes |
|-----------|----------------|-------|
| `pdf+zip` | suffix-append  | ZIP EOCD appended after `%%EOF`; central-directory offsets patched |
| `png+zip` | suffix-append  | ZIP EOCD appended after `IEND` chunk; offsets patched |
| `jpg+zip` | suffix-append  | ZIP EOCD appended after JPEG EOI marker; offsets patched |
| `mp3+zip` | suffix-append  | ZIP EOCD appended after last audio frame; offsets patched |
| `gif+sh`  | prefix-prepend | Shell script prepended; `exit 0` guard prevents shell parsing GIF binary |

---

## Installation

### Homebrew (macOS / Linux)

```sh
# coming soon
```

### Pre-built binaries

Download from the [releases page](../../releases) for Linux, macOS, or Windows (amd64/arm64).

### Build from source

```sh
git clone https://github.com/yourname/polyfile
cd polyfile
make build        # produces ./polyfile
make test         # run test suite
make lint         # vet + staticcheck
```

Requires Go 1.21+.

---

## Usage

### Build a polyglot

```sh
# PDF + ZIP
polyfile build document.pdf archive.zip -o combo.pdf

# PNG + ZIP
polyfile build image.png archive.zip -o combo.png

# JPEG + ZIP
polyfile build photo.jpg archive.zip -o combo.jpg

# MP3 + ZIP
polyfile build track.mp3 archive.zip -o combo.mp3

# GIF + shell script
polyfile build animation.gif script.sh -o combo.gif
bash combo.gif          # executes the shell script
```

### Verbose output (offsets + strategy info)

```sh
polyfile build document.pdf archive.zip -o combo.pdf --verbose
```

```
pair:      pdf + zip
strategy:  pdf+zip/suffix-append
input[0]:  document.pdf    pdf  45231 bytes
input[1]:  archive.zip     zip   8192 bytes

output:    combo.pdf
size:      53423 bytes
sha256:    a3f2...
```

### Inspect a file

```sh
polyfile inspect combo.pdf
```

```
file:  combo.pdf
size:  53423 bytes

FORMAT      STATUS
------      ------
pdf         valid
zip         valid

2 format(s) detected
```

### List supported pairs

```sh
polyfile list
```

### Format registry

```sh
polyfile formats list                        # show all registered handlers
polyfile formats show pdf                    # print handler details as TOML
polyfile formats validate combo.pdf --as zip # check magic bytes
```

---

## Flags

### `build`

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | — | Output file path (required) |
| `--verbose` | `-v` | false | Print strategy, inputs, offsets |
| `--verify` | | true | Validate output after writing |
| `--format` | | | Force format pair (e.g. `pdf+zip`) |
| `--strategy` | | | Force a specific strategy name |
| `--no-verify` | | | Skip output validation |

---

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Usage or logic error (unknown format, incompatible pair, bad flags) |
| 2 | I/O error (unreadable input, unwritable output) |
| 3 | Internal panic (bug) |

---

## License

MIT
