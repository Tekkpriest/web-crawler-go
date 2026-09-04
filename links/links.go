package links

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/net/html"
)

func Extract(r io.Reader, base url.URL) ([]*url.URL, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("error parsing html: %w", err)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if node.Type == html.ElementNode && node.Data == "a" {
			return nil, nil
		}
	} // checkt momentan nur die oberste Ebene, braucht noch rekursive Hilfsfunktion
	return nil, nil
}
