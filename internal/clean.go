package internal

import (
	"errors"
	"fmt"
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

func ConvertHtmlToTxtParse(content string, transpose int) (string, error) {
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

func ParseChord(chordString string) (Chord, error) {
	if len(chordString) == 0 {
		return Chord{}, errors.New("string parameter is empty")
	}
	var (
		rootOptions = make([]Note, 0)
		bassOptions = make([]Note, 0)
	)
	for _, note := range Notes {
		if strings.HasPrefix(chordString, note.String()) {
			rootOptions = append(rootOptions, note)
		}
	}
	if len(rootOptions) == 0 {
		return Chord{}, errors.New("no root was identified")
	}
	root := rootOptions[len(rootOptions)-1]
	if strings.Contains(chordString, "/") {
		for _, note := range Notes {
			if strings.HasSuffix(chordString, note.String()) {
				bassOptions = append(bassOptions, note)
			}
		}
	}
	chordStringWithoutRoot := strings.TrimPrefix(chordString, root.String())
	if len(bassOptions) != 0 {
		bass := bassOptions[len(bassOptions)-1]
		extension := strings.TrimSuffix(chordStringWithoutRoot, fmt.Sprintf("/%s", bass.String()))
		return CreateChordWithBass(root, extension, bass), nil
	} else {
		return CreateChord(root, chordStringWithoutRoot), nil
	}
}
