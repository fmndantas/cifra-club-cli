package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	Debug       bool
	Input       io.Reader
	Output      io.Writer
	ErrorOutput io.Writer
}

func (ctx *Context) input() io.Reader {
	if ctx.Input != nil {
		return ctx.Input
	}
	return os.Stdin
}

func (ctx *Context) output() io.Writer {
	if ctx.Output != nil {
		return ctx.Output
	}
	return os.Stdout
}

func (ctx *Context) errorOutput() io.Writer {
	if ctx.ErrorOutput != nil {
		return ctx.ErrorOutput
	}
	return os.Stderr
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
	results, err := searchSongs(http.DefaultClient, cmd.Content, cmd.Limit)
	if err != nil {
		return err
	}
	for i, song := range results {
		fmt.Fprintf(ctx.output(), "[%d] %s\n", i+1, fmt.Sprintf(printUrl, song.Dns, song.Url))
	}
	slog.Debug("done")
	return nil
}

func searchSongs(client *http.Client, content string, limit int) ([]SongResponse, error) {
	url := createSearchUrl(content, limit)
	slog.Debug("search url", "url", url)
	r, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searching %q: unexpected status %s", content, r.Status)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	slog.Debug("raw response", "body", string(body))
	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("when unmarshaling search result: %w", err)
	}
	slog.Debug("search completed", "number of results", len(result.Response.Docs))
	return result.Response.Docs, nil
}

func promptForSearchResult(input io.Reader, output io.Writer, results []string) (int, error) {
	for i, result := range results {
		fmt.Fprintf(output, "[%d] %s\n", i, result)
	}
	fmt.Fprint(output, "Select a result: ")
	line, err := bufio.NewReader(input).ReadString('\n')
	if err != nil && len(line) == 0 {
		return 0, err
	}
	index, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, fmt.Errorf("invalid search result index %q", strings.TrimSpace(line))
	}
	if index < 0 || index >= len(results) {
		return 0, fmt.Errorf("search result index %d is out of range (choose 0-%d)", index, len(results)-1)
	}
	return index, nil
}

type DownloadCmd struct {
	Url       string `arg:"" optional:"" name:"url" help:"Chord chart URL."`
	Query     string `help:"Search query; prompts for a result to download"`
	Limit     int    `help:"Maximum number of search results" default:"10" short:"l"`
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
	if cmd.Url != "" && cmd.Query != "" {
		return fmt.Errorf("provide either a URL or --query, not both")
	}
	if cmd.Url == "" && cmd.Query == "" {
		return fmt.Errorf("provide a URL or --query")
	}

	downloadURL := cmd.Url
	if cmd.Query != "" {
		results, err := searchSongs(http.DefaultClient, cmd.Query, cmd.Limit)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			return fmt.Errorf("no search results for %q", cmd.Query)
		}
		choices := make([]string, len(results))
		for i, song := range results {
			choices[i] = fmt.Sprintf(printUrl, song.Dns, song.Url)
		}
		index, err := promptForSearchResult(ctx.input(), ctx.errorOutput(), choices)
		if err != nil {
			return err
		}
		downloadURL = choices[index]
	}

	slog.Debug("download url", "url", downloadURL)
	chart, err := fetchChordChart(downloadURL, cmd.Transpose, internal.ConvertHtmlToTxtParse)
	if err != nil {
		return err
	}
	fmt.Fprint(ctx.output(), chart)
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
