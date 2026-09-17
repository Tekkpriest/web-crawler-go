package links

import (
	"net/url"
	"slices"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name string
		html string
		base string
		want []string
	}{
		{
			name: "absolute link",
			html: `<a href="https://go.dev/doc">d</a>`,
			base: "https://example.com/",
			want: []string{"https://go.dev/doc"},
		},
		{
			name: "relative to root",
			html: `<a href="/about">a</a>`,
			base: "https://example.com/team/",
			want: []string{"https://example.com/about"},
		},
		{
			name: "relative to current folder",
			html: `<a href="contact">c</a>`,
			base: "https://example.com/team/",
			want: []string{"https://example.com/team/contact"},
		},
		{
			name: "parent traversal",
			html: `<a href="../jobs">j</a>`,
			base: "https://example.com/team/list",
			want: []string{"https://example.com/jobs"},
		},
		{
			name: "non http schemata",
			html: `<a href="mailto:x@y.de">m</a><a href="tel:123">t</a><a href="/ok">o</a>`,
			base: "https://example.com/",
			want: []string{"https://example.com/ok"},
		},
		{
			name: "deeply nested",
			html: `<div><ul><li><a href="/deep">d</a></li></ul></div>`,
			base: "https://example.com/",
			want: []string{"https://example.com/deep"},
		},
		{
			name: "more than one link",
			html: `<a href="/one">1</a><a href="/two">2</a>`,
			base: "https://example.com/",
			want: []string{"https://example.com/one", "https://example.com/two"},
		},
		{
			name: "no links",
			html: `<p>nix hier</p>`,
			base: "https://example.com/",
			want: []string{},
		},
		{
			name: "a wihtout href",
			html: `<a name="anker">x</a>`,
			base: "https://example.com/",
			want: []string{},
		},
		{
			name: "fragment links",
			html: `<a href="#top">t</a>`,
			base: "https://example.com/index",
			want: []string{"https://example.com/index#top"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := url.Parse(tt.base)
			if err != nil {
				t.Fatalf("invalid base url in test: %v", err)
			}

			gotURLs, err := Extract(strings.NewReader(tt.html), baseURL)
			if err != nil {
				t.Fatalf("Extract() error: %v", err)
			}

			var got []string
			for _, u := range gotURLs {
				got = append(got, u.String())
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("Extract() = %v, want %v", got, tt.want)
			}
		})
	}
}
