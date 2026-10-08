package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fmndantas/cifraclubcli/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchSongs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "djavan lilas" {
			t.Errorf("query = %q, want %q", got, "djavan lilas")
		}
		if got := r.URL.Query().Get("limit"); got != "3" {
			t.Errorf("limit = %q, want %q", got, "3")
		}
		_ = json.NewEncoder(w).Encode(SearchResponse{Response: struct {
			Docs []SongResponse `json:"docs"`
		}{Docs: []SongResponse{{Dns: "djavan", Url: "lilas"}}}})
	}))
	defer srv.Close()

	oldSearchURL := searchUrl
	searchUrl = srv.URL + "/?%s"
	defer func() { searchUrl = oldSearchURL }()

	results, err := searchSongs(http.DefaultClient, "djavan lilas", 3)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "https://www.cifraclub.com.br/djavan/lilas/imprimir.html", fmt.Sprintf(printUrl, results[0].Dns, results[0].Url))
}

func TestPromptForSearchResult(t *testing.T) {
	var output bytes.Buffer
	result, err := promptForSearchResult(strings.NewReader("1\n"), &output, []string{"first", "second"})
	require.NoError(t, err)
	assert.Equal(t, 1, result)
	assert.Contains(t, output.String(), "[0] first")
	assert.Contains(t, output.String(), "[1] second")
}

func TestPromptForSearchResultRejectsOutOfRangeSelection(t *testing.T) {
	_, err := promptForSearchResult(strings.NewReader("2\n"), io.Discard, []string{"first", "second"})
	assert.EqualError(t, err, "search result index 2 is out of range (choose 0-1)")
}

func TestDownloadCmdRequiresExactlyOneSource(t *testing.T) {
	cmd := DownloadCmd{Url: "https://example.com", Query: "song"}
	err := cmd.Run(&Context{Output: io.Discard, ErrorOutput: io.Discard})
	assert.EqualError(t, err, "provide either a URL or --query, not both")
}

func TestDownloadCmdInteractiveSearchDownloadsSelectedResult(t *testing.T) {
	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"response":{"docs":[{"dns":"artist","url":"song"}]}}`)
	}))
	defer searchSrv.Close()

	chartSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "<html><body><pre data-chord-content=\"true\">Am</pre></body></html>")
	}))
	defer chartSrv.Close()

	oldSearchURL, oldPrintURL := searchUrl, printUrl
	searchUrl = searchSrv.URL + "/?%s"
	printUrl = chartSrv.URL + "/%s/%s"
	defer func() { searchUrl, printUrl = oldSearchURL, oldPrintURL }()

	var output, errorOutput bytes.Buffer
	cmd := DownloadCmd{Query: "artist song"}
	err := cmd.Run(&Context{
		Input:       strings.NewReader("0\n"),
		Output:      &output,
		ErrorOutput: &errorOutput,
	})
	require.NoError(t, err)
	assert.Equal(t, "Am\n", output.String())
	assert.Contains(t, errorOutput.String(), "[0]")
}

func TestFetchChordChart(t *testing.T) {
	var gotUserAgent string
	var gotAccept string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		gotPath = r.URL.Path
		w.Write([]byte("<html><body><pre data-chord-content=\"true\"><b>Am7</b>\nLara</pre></body></html>"))
	}))
	defer srv.Close()

	result, err := fetchChordChart(srv.URL+"/djavan/lilas/imprimir.html", 0, internal.ConvertHtmlToTxtParse)
	require.NoError(t, err)
	assert.Equal(t, "Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36", gotUserAgent, "server never saw the browser user-agent")
	assert.Equal(t, "*/*", gotAccept)
	assert.Equal(t, "/djavan/lilas/imprimir.html", gotPath)
	assert.Equal(t, "Am7\nLara\n", result, "response should arrive already cleaned")
}

func TestFetchChordChartRejectsNonOkStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetchChordChart(srv.URL, 0, internal.ConvertHtmlToTxtParse)
	require.Error(t, err)
}
