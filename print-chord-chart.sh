#!/bin/bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <cifra-url>" >&2
  echo "  e.g. $0 'https://www.cifraclub.com.br/djavan/lilas/'" >&2
  exit 1
fi

URL="$1"
UA='Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36'

# Prefer the print page: cleaner HTML, no wrappers inside <pre>, smaller download.
if [[ "$URL" != *imprimir.html ]]; then
  URL="${URL%/}/imprimir.html"
fi

curl -s --compressed -H "user-agent: $UA" "$URL" | python3 -c '
import re, html, sys
src = sys.stdin.read()
blocks = re.findall(r"<pre[^>]*data-chord-content[^>]*>(.*?)</pre>", src, re.S)
text = "\n".join(blocks) if blocks else src
text = re.sub(r"</?b[^>]*>", "", text)          # chord tags: drop, keep inner text
text = re.sub(r"</(?:div|span)>", "\n", text)   # closing wrappers -> newline
text = re.sub(r"<(?:div|span)[^>]*>", "", text) # opening wrappers -> drop
text = html.unescape(text)
text = re.sub(r"\n{3,}", "\n\n", text).strip()
print(text)
'
