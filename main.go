package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/fmndantas/cifraclubcli/internal"
)

var (
	searchUrl = "https://solr.sscdn.co/cc/c7/?%s"
	printUrl  = "https://www.cifraclub.com.br/%s/%s/imprimir.html"
	userAgent = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
)

func createSearchUrl(content string, limit int) string {
	params := url.Values{}
	params.Set("q", content)
	params.Set("limit", strconv.Itoa(limit))
	return fmt.Sprintf(searchUrl, params.Encode())
}

type Context struct {
	Debug bool
}

type SearchCmd struct {
	Limit   int    `help:"Maximum number of results the search should return" default:"10" short:"l"`
	Content string `arg:"" name:"content" help:"Content to search."`
}

type SongResponse struct {
	Dns string `json:"dns"`
	Url string `json:"url"`
}

type SearchResponse struct {
	Response struct {
		Docs []SongResponse `json:"docs"`
	} `json:"response"`
}

func (cmd *SearchCmd) Run(ctx *Context) error {
	slog.Debug("running search", "content", cmd.Content)
	url := createSearchUrl(cmd.Content, cmd.Limit)
	slog.Debug("search url", "url", url)
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	slog.Debug("raw response", "body", string(body))
	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("when unmarshaling search result: %w", err)
	}
	slog.Debug("search completed", "number of results", len(result.Response.Docs))
	for i, song := range result.Response.Docs {
		fmt.Printf("[%d] %s\n", i+1, fmt.Sprintf(printUrl, song.Dns, song.Url))
	}
	slog.Debug("done")
	return nil
}

type DownloadCmd struct {
	Url       string `arg:"" name:"url" help:"Chord chart URL."`
	Transpose int    `help:"Number of negative/positive semitones to transpose chords" default:"0" short:"t"`
}

// fetchChordChart downloads the print page and returns its cleaned plain-text
// chart, ending with a single newline (like the shell script's output).
func fetchChordChart(url string, transpose int, cleanFn func(string, int) (string, error)) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	slog.Debug("download response", "url", url, "status", r.Status)
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %s: unexpected status %s", url, r.Status)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return cleanFn(string(body), transpose)
}

func (cmd *DownloadCmd) Run(ctx *Context) error {
	slog.Info("running download")
	slog.Debug("download url", "url", cmd.Url)
	chart, err := fetchChordChart(cmd.Url, cmd.Transpose, internal.ConvertHtmlToTxtParse)
	if err != nil {
		return err
	}
	fmt.Print(chart)
	slog.Info("done")
	return nil
}

var cli struct {
	Debug    bool        `help:"Enable debug mode"`
	Search   SearchCmd   `cmd:"" help:"Search chord charts."`
	Download DownloadCmd `cmd:"" help:"Download a chord chart."`
}

func main() {
	ctx := kong.Parse(&cli)

	level := slog.LevelInfo
	if cli.Debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	err := ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
