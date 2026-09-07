package links

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func Extract(r io.Reader, base *url.URL) ([]*url.URL, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing html: %w", err)
	}

	var out []*url.URL
	walk(node, base, &out)

	return out, nil
}

func walk(node *html.Node, base *url.URL, out *[]*url.URL) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, a := range node.Attr {
			if a.Key == "href" {
				linkURL, err := url.Parse(a.Val)
				if err != nil {
					continue
				}

				absoluteURL := base.ResolveReference(linkURL)
				if absoluteURL.Scheme == "http" || absoluteURL.Scheme == "https" {
					*out = append(*out, absoluteURL)
				}
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		walk(c, base, out)
	}
}
