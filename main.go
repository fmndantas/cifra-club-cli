package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

var (
	searchUrl = "https://solr.sscdn.co/cc/c7/?q=%s&limit=%d"
	printUrl  = "https://www.cifraclub.com.br/%s/%s/imprimir.html"
)

func createSearchUrl(content string, limit int) string {
	fragments := strings.Split(content, " ")
	nonEmptyFragments := make([]string, 0)
	for _, f := range fragments {
		if len(f) > 0 {
			nonEmptyFragments = append(nonEmptyFragments, f)
		}
	}
	joinedContent := strings.Join(nonEmptyFragments, "+")
	return fmt.Sprintf(searchUrl, joinedContent, limit)
}

type Context struct {
	Debug bool
}

type SearchCmd struct {
	Limit   int    `help:"Maximum number of results the search should return"`
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
	slog.Info("running search")
	slog.Debug("search content", "content", cmd.Content)
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
		return err
	}
	slog.Info("search completed", "results", len(result.Response.Docs))
	for i, song := range result.Response.Docs {
		if len(song.Dns) > 0 && len(song.Url) > 0 {
			fmt.Printf("[%d] %s\n", i+1, fmt.Sprintf(printUrl, song.Dns, song.Url))
		}
	}
	return nil
}

var cli struct {
	Debug  bool      `help:"Enable debug mode"`
	Search SearchCmd `cmd:"" help:"Search chord charts."`
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
