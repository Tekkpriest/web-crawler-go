package links

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func Extract(r io.Reader, base *url.URL) ([]*url.URL, error) {
	if base == nil || !base.IsAbs() {
		return nil, fmt.Errorf("base url must be absolute, got: %v", base)
	}
	node, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	var found []*url.URL
	var walk func(*html.Node)

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}

				linkURL, err := url.Parse(a.Val)
				if err != nil {
					continue
				}

				absoluteURL := base.ResolveReference(linkURL)
				if absoluteURL.Scheme == "http" || absoluteURL.Scheme == "https" {
					found = append(found, absoluteURL)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(node)

	return found, nil
}
