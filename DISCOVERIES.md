# Cifra Club — discoveries: search & download cifras as plain text

Verified on 2026-09-23 against `https://www.cifraclub.com.br`. Test songs: "Lilás" by
Djavan and "Um Dia, Um Adeus" by Guilherme Arantes (id-suffixed URL:
`guilherme-arantes/um-dia-um-adeus-3571`). Both extracted completely with the same
deterministic algorithm. Example artifacts live in `example/`.

---

## 1. Search API

Cifra Club's website uses a Solr-backed search endpoint (no auth, no cookies, no special
headers — works with plain `curl`, even `curl/8.5.0` as UA):

```
https://solr.sscdn.co/cc/c7/?q=<query>&limit=<n>
```

- `q` — free text query, e.g. `djavan lilas` (artist + song works best; accents are optional)
- `limit` — max results
- The browser also sends `callback=suggest_callback` (JSONP), but it's optional — omit it
  and you get clean JSON.

### Example

```bash
curl 'https://solr.sscdn.co/cc/c7/?q=djavan+lilas&limit=10'
```

### Response structure

```json
{
  "response": {
    "numFound": 4,
    "docs": [
      {
        "art": "Djavan",          // artist name
        "dns": "djavan",          // artist slug -> URL path segment
        "txt": "Lilás",           // song name
        "url": "lilas",           // song slug -> URL path segment
        "id_song": 3794,
        "id_artist": 1417,
        "tipo": "2",              // "2" = cifra/song (what we want)
        "album_name": "Lilás",
        "url_album": "lilas-1988"
      }
    ]
  }
}
```

Notes:
- **A cifra's URL is `https://www.cifraclub.com.br/{dns}/{url}/`** — e.g.
  `djavan` + `lilas` → `https://www.cifraclub.com.br/djavan/lilas/`.
  Some songs have an id suffix in `url`, e.g. `um-dia-um-adeus-3571` — use it as-is.
- Filter docs that have **both `dns` and `url`** fields. Docs with `tipo: "6"` are albums,
  docs with `id_sb` are user songbooks/setlists — skip those.
- Results are already relevance-ranked (the API scores on text match, artist match,
  popularity and CTR), so the first cifra doc is usually the right one.

## 2. Downloading the cifra

### Anti-bot gotcha (Akamai)

`www.cifraclub.com.br` is behind Akamai and returns **403 Access Denied** for requests that
look like bots. Empirically (tested by dropping headers one by one):

| Request | Result |
|---|---|
| Browser UA only | 403 |
| `accept-encoding: gzip` + `curl/8.5.0` UA | 403 |
| Short UA (`Mozilla/5.0`) + accept-encoding | 403 |
| **Full Chrome UA + any `accept-encoding` header (even `identity`)** | **200** |

So the minimal working recipe is:

```bash
curl --compressed \
  -H 'user-agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36' \
  'https://www.cifraclub.com.br/djavan/lilas/'
```

(`--compressed` makes curl send `Accept-Encoding: gzip/deflate/br` and decompress automatically.)

### Best source: the print page

Two pages contain the cifra:

- `https://www.cifraclub.com.br/{dns}/{url}/` — main page, heavier HTML (ads, JS)
- `https://www.cifraclub.com.br/{dns}/{url}/imprimir.html` — **print version, much cleaner;
  use this one**

### HTML structure of the cifra

The site is a Next.js app. The cifra content lives in `<pre>` elements:

```html
<pre data-chord-content="true" data-chord-select="true">
  [Intro] <b data-chord-name="Am7(11)" ...>Am7(11)</b>  <b ...>F7M</b> ...
  E|-----3---3------0----- ...
</pre>
```

- The song is split across **multiple `<pre data-chord-content="true">` blocks** (Lilás has
  7, Um Dia Um Adeus has 2) — the count equals the print page count and varies per song;
  you must concatenate them all, in document order.
- Chords sit on a **separate "chord line" above the lyric line** (the classic two-line
  format). Each chord is a `<b data-chord-name="..." data-chord-original-text="...">…</b>`
  tag, and the chord name is the tag's text content (the attributes repeat the same name).
  The whitespace around the tags is what aligns each chord above the syllable/word where
  it's played.

  Example — a chord line and the lyric line that follows it:

  ```html
  <b data-chord-name="Am">Am</b>        <b data-chord-name="C7M">C7M</b>
  Eu quero ver o pôr do sol
  ```

  "Removing" here means **strip only the tag delimiters, keep the text between them**.
  Concretely, for the regex `</?b[^>]*>` (i.e. `<b ...>` and `</b>`), you *delete the match
  and keep everything else* — the chord name is the text *between* the opening and closing
  tags, so it survives untouched:

  ```text
  <b data-chord-name="Am">Am</b>
        \______removed______/\/\removed/
              (opening tag)     (closing tag)
            -> leaves just "Am"
  ```

  This is *not* the same as removing the whole `<b>…</b>` element: deleting the element
  would throw away the chord name too. A `re.sub(r"</?b[^>]*>", "", html)` does exactly
  the right thing — it matches only the tag, leaving the inner text in place.

  After applying that substitution, the two lines become:

  ```text
  Am        C7M
  Eu quero ver o pôr do sol
  ```

  `Am` and `C7M` stay on their own line, aligned above the words by the whitespace. Note
  there are **two lines** — the chord line and the lyric line — the alignment comes from
  the spaces, *not* from the chord being inline with the lyric.

  (The main page wraps each of these two-line groups in a `<div class="kvMV">`, and tab
  sections in `<div class="tabs"><span class="tab">…</span></div>`; the print page puts
  the same content inside `<pre data-chord-content>` blocks with plain newlines and **no**
  wrappers. The `<b>` chord structure is identical in both.)
- So the full plain-text extraction is: take all `pre[data-chord-content]` blocks →
  remove `<b>`/`</b>` tags → HTML-unescape entities → collapse 3+ newlines.
  This works as-is on the **print page** (no wrappers inside the `<pre>`s).

  **Main page caveat** (tested on `lilas.html`): the main page has a single `<pre>`
  holding the whole song, but every line group is wrapped in `<div class="kvMV">` (and
  tabs in `<span class="tab">`). The recipe above leaves those tags in the output, so you
  must also strip them — closing `</div>`/`</span>` become newlines, opening tags are
  dropped:

  ```python
  txt = re.sub(r'</?b[^>]*>', '', txt)          # chord tags: drop, keep inner text
  txt = re.sub(r'</(?:div|span)>', '\n', txt)   # closing wrappers -> newline
  txt = re.sub(r'<(?:div|span)[^>]*>', '', txt) # opening wrappers -> drop
  ```

  With that addition, the main page yields **exactly the same plain text** as the print
  page (verified: both normalize to identical 6662-char output for Lilás). The print page
  is still preferred — less to strip, smaller download.

  **Fragment caveat** (tested on `lilas-minified.html`, a bare `<div class="kvMV">` chunk
  with no `<pre>` wrapper): the pre-block regex finds **0 blocks** there. Fallback: if no
  `pre[data-chord-content]` match, apply the tag-stripping substitutions directly to the
  whole fragment/document. Same three regexes, applied without scoping — result is the
  same two-line chord chart.

### Metadata

The page embeds the song config in the (escaped) Next.js RSC JSON payload:

```
"config":{"capo":0,"keyShape":"C","tuning":"E A D G B E"}
```

- `capo` — capo fret (0 = none)
- `keyShape` — the song's key
- `tuning` — guitar tuning
- A `contributions` array (who transcribed it) follows it.

The old pages displayed "Tom: X" as text; the new UI only shows a transpose +/− control,
so parse `keyShape` out of the JSON if you need the key.

## 3. End-to-end working example

Search → pick first result → fetch print page → plain text:

```bash
#!/bin/bash
QUERY="djavan lilas"
UA='Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36'

# 1. search: first doc that has dns+url (first result for "djavan lilas" is Djavan - Lilás)
DOC=$(curl -s "https://solr.sscdn.co/cc/c7/?q=${QUERY// /+}&limit=30" \
      | python3 -c '
import sys, json
docs = json.load(sys.stdin)["response"]["docs"]
d = next(x for x in docs if x.get("dns") and x.get("url"))
print("{}/{}|{}|{}".format(d["dns"], d["url"], d["art"], d["txt"]))')
IFS="|" read -r PATH_ ARTIST SONG <<< "$DOC"

# 2. fetch print page (needs full browser UA + accept-encoding, else Akamai 403)
curl -s --compressed -H "user-agent: $UA" \
  "https://www.cifraclub.com.br/${PATH_}/imprimir.html" -o /tmp/cifra.html

# 3. extract plain text
python3 - <<'EOF'
import re, html
src = open('/tmp/cifra.html', encoding='utf-8').read()
blocks = re.findall(r'<pre[^>]*data-chord-content[^>]*>(.*?)</pre>', src, re.S)
text = '\n'.join(re.sub(r'</?b[^>]*>', '', b) for b in blocks)
text = html.unescape(text)
text = re.sub(r'\n{3,}', '\n\n', text).strip()
open('/tmp/cifra.txt', 'w', encoding='utf-8').write(text + '\n')
print(f'{len(text)} chars written')
EOF
```

Output (`/tmp/cifra.txt`) starts like:

```
[Intro] Am7(11)  F7M  Am7(11)  F7M
        Am7(11)  F7M  Am7(11)  F7M

[Dedilhado -Intro 4x]

   Am7(11)      F7M
E|-----3---3------0-------------------------|
B|-----3---3------1-----0h1-----------------|
...
```

## Summary

| Step | Endpoint | Requirements |
|---|---|---|
| Search | `https://solr.sscdn.co/cc/c7/?q=<query>&limit=<n>` | none (plain curl) |
| Cifra URL | `https://www.cifraclub.com.br/{dns}/{url}/` | — |
| Plain text | `.../{dns}/{url}/imprimir.html` | full browser UA + `Accept-Encoding` (else Akamai 403) |
| Parse | `pre[data-chord-content]` blocks, strip `<b>` tags, join, unescape | — |

### Validation

| Song | URL pattern | pre blocks | Output |
|---|---|---|---|
| Djavan — Lilás | `djavan/lilas` | 7 | `example/lilas-minified.html` (fragment), identical to print-page extraction |
| Guilherme Arantes — Um Dia, Um Adeus | `guilherme-arantes/um-dia-um-adeus-3571` (id suffix) | 2 | `example/guilherme-arantes-um-dia-um-adeus.txt` |

Algorithm is deterministic: pure regex + unescape rules, no randomness — same input
always yields identical output.
