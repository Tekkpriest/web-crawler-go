package crawl

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type want struct {
	statusCode int
	depth      int
	referrer   string
}

func TestCrawl(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><a href="/about">About</a></body></html>`))
	})
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<a href="/index">Index</a>
			<p><a href="/broken">Not Working</a></p>
			<a href="/jobs">Jobs</a>
			</body></html>`))
	})
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<a href="/about">About</a>
			<p><a href="/jobs/positioninfo.pdf">Download Job Offer as .pdf</a></p>
			</body></html>`))
	})
	mux.HandleFunc("/jobs/positioninfo.pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte(``))
	})

	testserver := httptest.NewServer(mux)
	defer testserver.Close()

	u, err := url.Parse(testserver.URL)
	if err != nil {
		t.Fatalf("parsing test server: %v", err)
	}
	host := u.Hostname()

	c := &Crawler{
		Client:    testserver.Client(),
		Host:      host,
		MaxDepth:  3,
		UserAgent: "test-agent",
	}

	start, err := url.Parse(testserver.URL + "/index")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}

	results := c.Crawl(context.Background(), start)
	gotURL := make(map[string]Result, len(results))

	for _, r := range results {
		gotURL[r.URL] = r
	}

	wanted := map[string]want{
		testserver.URL + "/index":                 {200, 0, ""},
		testserver.URL + "/about":                 {200, 1, testserver.URL + "/index"},
		testserver.URL + "/broken":                {404, 2, testserver.URL + "/about"},
		testserver.URL + "/jobs":                  {200, 2, testserver.URL + "/about"},
		testserver.URL + "/jobs/positioninfo.pdf": {200, 3, testserver.URL + "/jobs"},
	}

	if len(results) != len(wanted) {
		t.Errorf("got %d results, wanted %d", len(results), len(wanted))
	}

	for wantKey, wantURL := range wanted {
		got, ok := gotURL[wantKey]
		if !ok {
			t.Errorf("missing result for %q", wantKey)
			continue
		}

		if got.StatusCode != wantURL.statusCode {
			t.Errorf("got status code %d, wanted %d", got.StatusCode, wantURL.statusCode)
		}

		if got.Depth != wantURL.depth {
			t.Errorf("got depth %d, wanted %d", got.Depth, wantURL.depth)
		}

		if got.Referrer != wantURL.referrer {
			t.Errorf("got referrer %s, wanted %s", got.Referrer, wantURL.referrer)
		}
	}
}

func TestNormalise(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "remove fragment",
			url:  "https://example.com/index#top",
			want: "https://example.com/index",
		},
		{
			name: "remove trailing slash",
			url:  "https://example.com/about/",
			want: "https://example.com/about",
		},
		{
			name: "root slash staying",
			url:  "https://example.com/",
			want: "https://example.com/",
		},
		{
			name: "no change",
			url:  "https://example.com/about",
			want: "https://example.com/about",
		},
		{
			name: "query parameters staying",
			url:  "https://example.com/about?darkmode=1#top",
			want: "https://example.com/about?darkmode=1",
		},
		{
			name: "remove fragment & trailing slash",
			url:  "https://example.com/about/#customer",
			want: "https://example.com/about",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("error parsing (%q): %v", tt.url, err)
			}

			got := normalise(u)
			if got != tt.want {
				t.Errorf("normalise (%q) = %q, expected %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsSameHost(t *testing.T) {
	tests := []struct {
		name string
		url  string
		host string
		want bool
	}{
		{
			name: "exact match",
			url:  "https://example.com/",
			host: "example.com",
			want: true,
		},
		{
			name: "real subdomain",
			url:  "https://blog.example.com/",
			host: "example.com",
			want: true,
		},
		{
			name: "multi subdomain",
			url:  "https://blog.internal.example.com/",
			host: "example.com",
			want: true,
		},
		{
			name: "fake subdomain",
			url:  "https://evilexample.com/",
			host: "example.com",
			want: false,
		}, {
			name: "different domain",
			url:  "https://go.dev/doc",
			host: "example.com",
			want: false,
		}, {
			name: "upper/lowercase",
			url:  "https://Example.com/",
			host: "example.com",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Fatalf("error parsing (%q): %v", tt.url, err)
			}

			got := isSameHost(u, tt.host)
			if got != tt.want {
				t.Errorf("isSameHost (%q) = %v, expected %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsBroken(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name:   "success 200",
			result: Result{StatusCode: 200},
			want:   false,
		},
		{
			name:   "redirect 301",
			result: Result{StatusCode: 301},
			want:   false,
		},
		{
			name:   "client error 404",
			result: Result{StatusCode: 404},
			want:   true,
		},
		{
			name:   "server error 500",
			result: Result{StatusCode: 500},
			want:   true,
		},
		{
			name:   "network error with status 0",
			result: Result{Err: errors.New("connection refused"), StatusCode: 0},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsBroken(); got != tt.want {
				t.Errorf("IsBroken() = %v, want %v", got, tt.want)
			}
		})
	}
}
