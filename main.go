package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/tekkpriest/web-crawler-go/crawl"
)

func main() {
	os.Exit(run())
}

func run() int {
	urlInput := flag.String("url", "", "url to start crawling from")
	depth := flag.Int("depth", 2, "maximum crawl depth, default 2")
	timeout := flag.Duration("timeout", 10*time.Second, "http request timeout, default 10s")
	outputJSON := flag.Bool("json", false, "json format output for use with other tools/scripts")
	outputVerbose := flag.Bool("v", false, "full history for all links")

	flag.Parse()

	if *urlInput == "" {
		fmt.Fprintln(os.Stderr, "error: url is needed")
		flag.Usage()
		return 2
	}

	startURL, err := url.Parse(*urlInput)
	if err != nil || !startURL.IsAbs() {
		fmt.Fprintln(os.Stderr, "error: invalid start url")
		flag.Usage()
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c := &crawl.Crawler{
		Client: &http.Client{
			Timeout: *timeout,
		},
		Host:      startURL.Hostname(),
		MaxDepth:  *depth,
		UserAgent: "web-crawler-go/1.0",
	}

	results := c.Crawl(ctx, startURL)

	switch {
	case *outputJSON:
		if err := writeJSON(os.Stdout, results); err != nil {
			fmt.Fprintln(os.Stderr, "error: could not encode json")
			return 1
		}
	case *outputVerbose:
		writeVerbose(os.Stdout, results)
	default:
		writeReport(os.Stdout, results)
	}

	for _, r := range results {
		if r.IsBroken() {
			return 1
		}
	}

	return 0
}
