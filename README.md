# cifraclubcli

A small Go CLI to search and download chord charts (cifras) from
[Cifra Club](https://www.cifraclub.com.br).

Built for automation: both commands are scriptable and machine-friendly, so they
compose well in scripts, cron jobs, and pipelines.

## Install

```bash
go build -o cifraclubcli .
```

## Usage

### Search

```bash
cifraclubcli search "djavan lilas" --limit=5
```

Prints matching songs with their `imprimir.html` URLs, one per line:

```
[1] https://www.cifraclub.com.br/djavan/lilas/imprimir.html
```

### Download

```bash
cifraclubcli download "https://www.cifraclub.com.br/djavan/lilas/imprimir.html" > lilas.txt
```

Fetches the print page and writes the cleaned chord chart as plain text to stdout.

Transpose all chords by N semitones (negative = down) with `--transpose`:

```bash
cifraclubcli download "https://www.cifraclub.com.br/djavan/lilas/imprimir.html" --transpose=2
```

### Options

| Flag | Command | Description |
|---|---|---|
| `--limit`, `-l` | search | Maximum number of results (default: 10) |
| `--transpose`, `-t` | download | Semitones to transpose chords (default: 0) |
| `--debug` | global | Debug logging (search URL, raw API response, etc.) |

## Layout

```
.
├── main.go              # CLI entrypoint (search + download commands)
├── internal/            # HTML→text cleaning, chord parsing, transposition
├── print-chord-chart.sh # reference script: fetch + extract a cifra as plain text
├── DISCOVERIES.md       # research notes on the Cifra Club API/HTML
└── examples/            # saved HTML + extracted text for test songs
```

## How it works

See [DISCOVERIES.md](DISCOVERIES.md) for details on the search API
(`solr.sscdn.co`), the anti-bot rules for downloading, and chart extraction.

## Limitations

- `--transpose` only transposes chords; tablatures (tabs) are not transposed.

## Tests

```bash
make test   # or: go test ./...
```
