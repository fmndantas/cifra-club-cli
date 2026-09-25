package internal

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func showParentInfo(node *html.Node) string {
	if node == nil {
		return "[parent is nil]"
	}
	return fmt.Sprintf("parent[type=%s, data=%s]", node.Type.String(), node.Data)
}

func exploreHierarchy(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "pre" {
		for bar := range node.Descendants() {
			if (bar.Type == html.ElementNode && bar.Data == "b") || bar.Type == html.TextNode {
				fmt.Printf(
					"  data=%q, attr=%v, type=%q, %s\n",
					bar.Data,
					bar.Attr,
					bar.Type, showParentInfo(bar.Parent),
				)
			}
		}
	}
}

func TestExploreNetHttpParse(t *testing.T) {
	t.Skip("learning test")
	f, _ := os.Open("../examples/lilas.html")
	node, err := html.Parse(f)
	require.NoError(t, err, "parse")

	for foo := range node.Descendants() {
		exploreHierarchy(foo)
	}
}
