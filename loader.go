package artisum

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func ExtractTextContentFromURL(url string) (string, error) {
	h, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer h.Body.Close()

	body, err := io.ReadAll(h.Body)
	if err != nil {
		return "", err
	}

	parsed, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	buf := &strings.Builder{}

	var f func(*html.Node) error
	f = func(n *html.Node) error {
		if n.Type == html.TextNode && n.Parent.Data != "script" && n.Parent.Data != "style" {
			if _, err := buf.WriteString(strings.TrimSpace(n.Data)); err != nil {
				return err
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err = f(c); err != nil {
				return err
			}
		}
		return nil
	}
	f(parsed)

	return buf.String(), nil
}
