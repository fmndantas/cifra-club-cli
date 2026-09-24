# cifra-club-cli

A small Go CLI to find and download chord charts (cifras) from
[Cifra Club](https://www.cifraclub.com.br).

**One of the purposes of this project is to enable automations.** It is built
as a scriptable, machine-friendly CLI so the search and download steps can be
composed into scripts, cron jobs, and pipelines — not a human-only interactive
tool.

## Install

```bash
go build -o cifra-club-cli .
```

## Usage

### Search

```bash
cifra-club-cli search "djavan lilas" --limit=5
```

Prints the matching songs and their `imprimir.html` URLs, one per line:

```
[1] https://www.cifraclub.com.br/djavan/lilas/imprimir.html
```

Options:

| Flag | Description |
|---|---|
| `--limit` | Maximum number of results (defaults to 0 = server default) |
| `--debug` | Enable debug logging (prints search URL, raw JSON response, etc.) |

### Download

```bash
cifra-club-cli download "https://www.cifraclub.com.br/djavan/lilas/" --output lilas.txt
```

> Note: `download` is a work-in-progress stub and does not save a file yet.
> Use `print-chord-chart.sh` in the meantime.

## Debug logging

`--debug` bumps `log/slog` from `INFO` to `DEBUG`:

```bash
cifra-club-cli --debug search "djavan lilas"
```

## How it works

See [DISCOVERIES.md](DISCOVERIES.md) for the reverse-engineered details, including:

- The search API (`solr.sscdn.co`) and its response shape.
- The Akamai anti-bot rules for downloading cifras (full browser UA + `Accept-Encoding`).
- How to extract plain-text chord charts from the print page (`imprimir.html`).

## Layout

```
.
├── main.go                 # CLI entrypoint (kong), search + download commands
├── main_test.go            # tests
├── print-chord-chart.sh    # reference script: fetch + extract a cifra as plain text
├── DISCOVERIES.md          # research notes on the Cifra Club API/HTML
└── examples/               # saved HTML + extracted text for test songs
```

## Tests

```bash
go test ./...
```
