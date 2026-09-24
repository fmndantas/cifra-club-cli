package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	result, err := fetchChordChart(srv.URL + "/djavan/lilas/imprimir.html")
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

	_, err := fetchChordChart(srv.URL)
	require.Error(t, err)
}
