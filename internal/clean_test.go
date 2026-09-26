package internal_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fmndantas/cifraclubcli/internal"
)

func TestConvertHtmlToTxtRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "extracts and cleans chord content blocks",
			input: "<pre data-chord-content=\"true\"><b>C</b>&amp;\n\n\nX</pre><pre data-chord-content>\n<span>Y</span></pre>",
			want:  "C&\n\nX\n\nY\n",
		},
		{
			name:  "cleans the whole input when there are no content blocks",
			input: "<div><b>A</b></div><span>B</span>&#x1F3B8;\n\n\n",
			want:  "A\nB\n🎸\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := internal.ConvertHtmlToTxtRegex(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestChordChartConversionWithRegex(t *testing.T) {
	t.Skip("deprecated")
	cases := []struct {
		id           string
		htmlFile     string
		expectedFile string
	}{
		{"lilas", "lilas.html", "lilas.txt"},
		{"um dia um adeus", "um-dia-um-adeus.html", "um-dia-um-adeus.txt"},
	}
	for _, tt := range cases {
		t.Run(tt.id, func(t *testing.T) {
			htmlContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.htmlFile))
			require.NoError(t, err, "read html file")
			expectedContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.expectedFile))
			require.NoError(t, err, "read expected file")
			result, err := internal.ConvertHtmlToTxtRegex(string(htmlContent))
			require.NoError(t, err)
			assert.Equal(t, string(expectedContent), result, "result is not the expected")
		})
	}
}

func TestChordChartConversionWithParse(t *testing.T) {
	cases := []struct {
		id           string
		transpose    int
		htmlFile     string
		expectedFile string
	}{
		{"lilas", 0, "lilas.html", "lilas.txt"},
		{"um dia um adeus", 0, "um-dia-um-adeus.html", "um-dia-um-adeus.txt"},
		{"um dia um adeus", 1, "um-dia-um-adeus.html", "um-dia-um-adeus-1-above.txt"},
	}
	for _, tt := range cases {
		t.Run(tt.id, func(t *testing.T) {
			htmlContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.htmlFile))
			require.NoError(t, err, "read html file")
			expectedContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.expectedFile))
			require.NoError(t, err, "read expected file")
			result, err := internal.ConvertHtmlToTxtParse(string(htmlContent), tt.transpose)
			require.NoError(t, err)
			assert.Equal(t, string(expectedContent), result, "result is not the expected")
		})
	}
}

func TestChordParsing(t *testing.T) {
	cases := []struct {
		id            int
		stringValue   string
		expectedChord internal.Chord
	}{
		{1, "C", internal.CreateChord(internal.C, "")},
		{2, "Cadd9", internal.CreateChord(internal.C, "add9")},
		{3, "C/E", internal.CreateChordWithBass(internal.C, "", internal.E)},
		{4, "C#M7/Bb", internal.CreateChordWithBass(internal.CSharp, "M7", internal.BFlat)},
		{5, "F#7(#9/b9/#5)/G", internal.CreateChordWithBass(internal.FSharp, "7(#9/b9/#5)", internal.G)},
		{6, "F#7(#9/b9/#5)", internal.CreateChord(internal.FSharp, "7(#9/b9/#5)")},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("case-%d", tt.id), func(t *testing.T) {
			result, err := internal.ParseChord(tt.stringValue)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedChord.Root, result.Root, "root")
			assert.Equal(t, tt.expectedChord.Extension, result.Extension, "extension")
			assert.Equal(t, tt.expectedChord.HasBass(), result.HasBass(), "has bass")
			if tt.expectedChord.HasBass() {
				assert.Equal(t, *tt.expectedChord.Bass, *result.Bass, "bass")
			}
		})
	}
}
