package crawl

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/tekkpriest/web-crawler-go/links"
)

type Result struct {
	URL        string
	StatusCode int
	Err        error
	Depth      int
	Referrer   string
}

type Crawler struct {
	Client    *http.Client
	Host      string
	MaxDepth  int
	UserAgent string
}

type item struct {
	url      *url.URL
	depth    int
	referrer string
}

func (c *Crawler) Crawl(ctx context.Context, start *url.URL) []Result {
	var results []Result
	visited := map[string]bool{}
	queue := []item{{url: start}}

	for len(queue) > 0 {

		select {
		case <-ctx.Done():
			return results
		default:
		}

		current := queue[0]
		queue = queue[1:]

		key := normalise(current.url)

		if visited[key] {
			continue
		}

		visited[key] = true

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, current.url.String(), nil)
		if err != nil {
			results = append(results, Result{
				URL:      key,
				Err:      err,
				Depth:    current.depth,
				Referrer: current.referrer,
			})
			continue
		}

		req.Header.Set("User-Agent", c.UserAgent)
		resp, err := c.Client.Do(req)
		if err != nil {
			results = append(results, Result{
				URL:      key,
				Err:      err,
				Depth:    current.depth,
				Referrer: current.referrer,
			})
			continue
		}

		results = append(results, Result{
			URL:        key,
			StatusCode: resp.StatusCode,
			Depth:      current.depth,
			Referrer:   current.referrer,
		})

		func() {
			defer resp.Body.Close()

			mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
			if err == nil && isSameHost(current.url, c.Host) && mediaType == "text/html" && current.depth < c.MaxDepth {
				limited := io.LimitReader(resp.Body, 10<<20)

				found, err := links.Extract(limited, resp.Request.URL)
				if err == nil {
					for _, link := range found {
						queue = append(queue, item{
							url:      link,
							depth:    current.depth + 1,
							referrer: key,
						})
					}
				}
			}
		}()
	}

	return results
}

func normalise(u *url.URL) string {
	cp := *u
	cp.Fragment = ""

	if cp.Path != "/" {
		cp.Path = strings.TrimSuffix(cp.Path, "/")
	}

	return cp.String()
}

func isSameHost(u *url.URL, host string) bool {
	checkHost := strings.ToLower(u.Hostname())
	host = strings.ToLower(host)

	return checkHost == host || strings.HasSuffix(checkHost, "."+host)
}

func (r Result) IsBroken() bool {
	return r.Err != nil || r.StatusCode >= 400
}
