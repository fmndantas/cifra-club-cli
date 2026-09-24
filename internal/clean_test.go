package internal_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fmndantas/cifraclubcli/internal"
)

func TestCleanHtmlFile(t *testing.T) {
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
			assert.Equal(t, tt.want, internal.CleanHtmlFile(tt.input))
		})
	}
}

func TestCleanPrintedChordChart(t *testing.T) {
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
			result := internal.CleanHtmlFile(string(htmlContent))
			assert.Equal(t, string(expectedContent), result, "result is not the expected")
		})
	}
}
