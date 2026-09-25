package internal

import (
	"html"
	"regexp"
	"strings"
)

func ConvertHtmlChordChartToTxt(content string) string {
	var (
		chordContentRE = regexp.MustCompile(`(?s)<pre[^>]*data-chord-content[^>]*>(.*?)</pre>`)
		bTagRE         = regexp.MustCompile(`</?b[^>]*>`)
		closingTagRE   = regexp.MustCompile(`</(?:div|span)>`)
		openingTagRE   = regexp.MustCompile(`<(?:div|span)[^>]*>`)
		newlinesRE     = regexp.MustCompile(`\n{3,}`)
	)
	blocks := chordContentRE.FindAllStringSubmatch(content, -1)
	text := content
	if len(blocks) > 0 {
		parts := make([]string, len(blocks))
		for i, block := range blocks {
			parts[i] = block[1]
		}
		text = strings.Join(parts, "\n")
	}

	text = bTagRE.ReplaceAllString(text, "")
	text = closingTagRE.ReplaceAllString(text, "\n")
	text = openingTagRE.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = newlinesRE.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text) + "\n"
}
