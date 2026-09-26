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
		{"um dia um adeus 1 above", 1, "um-dia-um-adeus.html", "um-dia-um-adeus-1-above.txt"},
	}
	for _, tt := range cases {
		t.Run(tt.id, func(t *testing.T) {
			htmlContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.htmlFile))
			require.NoError(t, err, "read html file")
			expectedContent, err := os.ReadFile(fmt.Sprintf("../examples/%s", tt.expectedFile))
			require.NoError(t, err, "read expected file")
			result, err := internal.ConvertHtmlToTxtParse(string(htmlContent), tt.transpose)
			require.NoError(t, err)
			assert.Equal(
				t,
				string(expectedContent),
				result,
				"result is not the expected",
			)
		})
	}
}

func TestChordParsing(t *testing.T) {
	cases := []struct {
		id            int
		expectedChord internal.Chord
		stringValue   string
	}{
		{1, internal.CreateChord(internal.C, ""), "C"},
		{2, internal.CreateChord(internal.C, "add9"), "Cadd9"},
		{3, internal.CreateChordWithBass(internal.C, "", internal.E), "C/E"},
		{4, internal.CreateChordWithBass(internal.CSharp, "M7", internal.BFlat), "C#M7/Bb"},
		{5, internal.CreateChordWithBass(internal.FSharp, "7(#9/b9/#5)", internal.G), "F#7(#9/b9/#5)/G"},
		{6, internal.CreateChord(internal.FSharp, "7(#9/b9/#5)"), "F#7(#9/b9/#5)"},
		{7, internal.CreateChordWithBass(internal.AFlat, "", internal.C), "Ab/C"},
		{8, internal.CreateChordWithBass(internal.EFlat, "", internal.G), "Eb/G"},
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

func TestRespaceChord(t *testing.T) {
	cases := []struct {
		id                             int
		originalChord, transposedChord internal.Chord
		expectedResult                 string
	}{
		{1, internal.CreateChord(internal.C, ""), internal.CreateChord(internal.C, ""), "C"},
		{2, internal.CreateChord(internal.C, ""), internal.CreateChord(internal.CSharp, ""), "C#@"},
		{3, internal.CreateChord(internal.CSharp, ""), internal.CreateChord(internal.C, ""), "C "},
		{4, internal.CreateChord(internal.CSharp, "m7"), internal.CreateChord(internal.D, "m7"), "Dm7 "},
		{
			5,
			internal.CreateChordWithBass(internal.FSharp, "", internal.ASharp),
			internal.CreateChordWithBass(internal.G, "", internal.B),
			"G/B  ",
		},
		{
			6,
			internal.CreateChordWithBass(internal.F, "", internal.A),
			internal.CreateChordWithBass(internal.FSharp, "", internal.ASharp),
			"F#/A#$",
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("case-%d", tt.id), func(t *testing.T) {
			result := internal.RespaceChord(tt.originalChord, tt.transposedChord)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestRespaceChordChartLine(t *testing.T) {
	cases := []struct {
		id                           int
		originalText, expectedResult string
	}{
		{1, "Cm7@  Dm\n", "Cm7 Dm\n"},
		{2, "Cm7@ Dm\n", "Cm7 Dm\n"},
		{3, "Cm7@ Dm Em\n", "Cm7 Dm Em\n"}, // test if shouldRemove resets between chords
		{4, "Cm7 Dm Em@\n", "Cm7 Dm Em\n"},
		{5, "Cm7 Dm Em$\n", "Cm7 Dm Em\n"},
		{6, "Cm7 Dm$ Em\n", "Cm7 Dm Em\n"},
		{7, "Cm7 Dm$  Em\n", "Cm7 Dm Em\n"},
		{8, "Cm7 Dm$   Em\n", "Cm7 Dm Em\n"},
		{9, "Cm7 Dm$   Em Cm\n", "Cm7 Dm Em Cm\n"},
		{10, "Cm7 Dm$    Em Cm\n", "Cm7 Dm  Em Cm\n"},
		{10, "Cm7$Dm   Em Cm\n", "Cm7Dm   Em Cm\n"},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("case-%d", tt.id), func(t *testing.T) {
			result, err := internal.RespaceChordChartLine(tt.originalText)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
