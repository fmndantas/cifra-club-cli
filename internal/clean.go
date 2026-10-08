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
	var sb strings.Builder
	for node := range document.Descendants() {
		if node.Type != netHtml.ElementNode || node.Data != "pre" {
			continue
		}
		for d := range node.Descendants() {
			if d.Type != netHtml.TextNode {
				continue
			}
			isThisNodeAChord := d.Parent != nil && d.Parent.Type == netHtml.ElementNode && d.Parent.Data == "b"
			if isThisNodeAChord {
				originalChord, err := ParseChord(d.Data)
				if err != nil {
					return "", err
				}
				transposedChord := originalChord.Transpose(transpose)
				respacedChord, err := RespaceChord(originalChord, transposedChord)
				if err != nil {
					return "", err
				}
				sb.WriteString(respacedChord)
			} else {
				sb.WriteString(d.Data)
			}
		}
		sb.WriteString("\n")
	}
	sbResult := sb.String()
	sb.Reset()
	for line := range strings.Lines(sbResult) {
		respacedLine, err := RespaceChordChartLine(line)
		if err != nil {
			return "", err
		} else {
			sb.WriteString(respacedLine)
		}
	}
	return sb.String(), nil
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

func RespaceChord(originalChord, transposedChord Chord) (string, error) {
	var (
		originalRepr      = originalChord.String()
		transposedRepr    = transposedChord.String()
		lenOriginalRepr   = len(originalRepr)
		lenTransposedRepr = len(transposedRepr)
	)
	switch {
	case lenOriginalRepr == lenTransposedRepr:
		return transposedRepr, nil
	case lenOriginalRepr+1 == lenTransposedRepr:
		return transposedRepr + string(removeOneSpaceRune), nil
	case lenOriginalRepr+2 == lenTransposedRepr:
		return transposedRepr + string(removeTwoSpacesRune), nil
	case lenOriginalRepr > lenTransposedRepr:
		return transposedRepr + strings.Repeat(" ", lenOriginalRepr-lenTransposedRepr), nil
	default:
		return "", fmt.Errorf(
			"unexpected case in chord respacing. originalRepr=%s, transposedRepr=%s",
			originalRepr,
			transposedRepr,
		)
	}
}

func RespaceChordChartLine(line string) (string, error) {
	if !strings.HasSuffix(line, "\n") {
		return "", errors.New("line does not ends with \"\n\"")
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
