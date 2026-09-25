package internal

import (
	"html"
	"regexp"
	"strings"

	netHtml "golang.org/x/net/html"
)

func ConvertHtmlToTxtRegex(content string) (string, error) {
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

	return strings.TrimSpace(text) + "\n", nil
}

func ConvertHtmlToTxtParse(content string) (string, error) {
	document, err := netHtml.Parse(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	var result strings.Builder
	for node := range document.Descendants() {
		if node.Type == netHtml.ElementNode && node.Data == "pre" {
			for d := range node.Descendants() {
				if d.Type == netHtml.TextNode {
					result.WriteString(d.Data)
				}
			}
			result.WriteString("\n")
		}
	}
	return result.String(), nil
}
