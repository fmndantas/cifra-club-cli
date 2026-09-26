package internal

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strings"

	netHtml "golang.org/x/net/html"
)

const (
	removeOneSpaceRune  = '@'
	removeTwoSpacesRune = '$'
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
	var (
		sb   strings.Builder
		errs error
	)
	for node := range document.Descendants() {
		if node.Type == netHtml.ElementNode && node.Data == "pre" {
			for d := range node.Descendants() {
				isThisNodeAChord := d.Parent != nil && d.Parent.Type == netHtml.ElementNode && d.Parent.Data == "b"
				if d.Type == netHtml.TextNode {
					if isThisNodeAChord {
						originalChord, err := ParseChord(d.Data)
						transposedChord := originalChord.Transpose(transpose)
						if err != nil {
							errs = errors.Join(err)
						} else {
							sb.WriteString(RespaceChord(originalChord, transposedChord))
						}
					} else {
						sb.WriteString(d.Data)
					}
				}
			}
			sb.WriteString("\n")
		}
	}
	sbResult := sb.String()
	sb.Reset()
	for line := range strings.Lines(sbResult) {
		respacedLine, err := RespaceChordChartLine(line)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			sb.WriteString(respacedLine)
		}
	}
	if errs != nil {
		return "", errs
	} else {
		return sb.String(), nil
	}
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
	getMostLenghtyNote := func(options []Note) (Note, error) {
		if len(options) == 0 {
			return 0, errors.New("getLenghtyNote: no options were available")
		}
		return slices.MaxFunc(
			options,
			func(n1 Note, n2 Note) int {
				if len(n1.String()) > len(n2.String()) {
					return 1
				} else {
					return -1
				}
			},
		), nil
	}
	root, err := getMostLenghtyNote(rootOptions)
	if err != nil {
		return Chord{}, err
	}
	chordStringWithoutRoot := strings.TrimPrefix(chordString, root.String())
	for _, note := range Notes {
		if strings.HasSuffix(chordStringWithoutRoot, note.String()) {
			bassOptions = append(bassOptions, note)
		}
	}
	if len(bassOptions) != 0 {
		bass, err := getMostLenghtyNote(bassOptions)
		if err != nil {
			return Chord{}, err
		}
		extension := strings.TrimSuffix(chordStringWithoutRoot, fmt.Sprintf("/%s", bass.String()))
		return CreateChordWithBass(root, extension, bass), nil
	} else {
		return CreateChord(root, chordStringWithoutRoot), nil
	}
}

// TODO: add error to the return
func RespaceChord(originalChord, transposedChord Chord) string {
	var (
		originalRepr      = originalChord.String()
		transposedRepr    = transposedChord.String()
		lenOriginalRepr   = len(originalRepr)
		lenTransposedRepr = len(transposedRepr)
	)
	switch {
	case lenOriginalRepr == lenTransposedRepr:
		return transposedRepr
	case lenOriginalRepr+1 == lenTransposedRepr:
		return transposedRepr + string(removeOneSpaceRune)
	case lenOriginalRepr+2 == lenTransposedRepr:
		return transposedRepr + string(removeTwoSpacesRune)
	case lenOriginalRepr > lenTransposedRepr:
		return transposedRepr + strings.Repeat(" ", lenOriginalRepr-lenTransposedRepr)
	default:
		panic("TODO")
	}
}

func RespaceChordChartLine(line string) (string, error) {
	if !strings.HasSuffix(line, "\n") {
		panic("TODO")
	}
	var (
		shouldRemoveOneSpace  = false
		shouldRemoveTwoSpaces = false
		foundFirstSpace       = false
		foundSecondSpace      = false
		sb                    strings.Builder
	)
	for _, c := range line {
		switch {
		case c == removeOneSpaceRune:
			shouldRemoveOneSpace = true
			shouldRemoveTwoSpaces = false
			foundFirstSpace = false
			foundSecondSpace = false
		case c == removeTwoSpacesRune:
			shouldRemoveOneSpace = false
			shouldRemoveTwoSpaces = true
			foundFirstSpace = false
			foundSecondSpace = false
		case shouldRemoveOneSpace && c == ' ':
			if !foundFirstSpace {
				sb.WriteRune(c)
				foundFirstSpace = true
			} else {
				shouldRemoveOneSpace = false
				foundFirstSpace = false
			}
		case shouldRemoveTwoSpaces && c == ' ':
			if !foundFirstSpace {
				sb.WriteRune(c)
				foundFirstSpace = true
			} else if !foundSecondSpace {
				foundSecondSpace = true
			} else {
				shouldRemoveTwoSpaces = false
				foundSecondSpace = false
			}
		case (shouldRemoveOneSpace || shouldRemoveTwoSpaces) && c != ' ' && c != removeOneSpaceRune && c != removeTwoSpacesRune:
			sb.WriteRune(c)
			shouldRemoveOneSpace = false
			shouldRemoveTwoSpaces = false
			foundFirstSpace = false
			foundSecondSpace = false
		default:
			sb.WriteRune(c)
		}
	}
	return strings.TrimRightFunc(sb.String(), func(r rune) bool { return r == '\n' || r == ' ' }) + "\n", nil
}
