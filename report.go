package main

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/tekkpriest/web-crawler-go/crawl"
)

func writeJSON(w io.Writer, results []crawl.Result) error {
	return json.NewEncoder(w).Encode(results)
}

func writeReport(w io.Writer, results []crawl.Result) {
	var total, ok, broken int

	brokenByRef := make(map[string][]crawl.Result)
	for _, r := range results {
		total++
		if r.IsBroken() {
			broken++
			ref := r.Referrer
			if ref == "" {
				ref = "(start url)"
			}
			brokenByRef[ref] = append(brokenByRef[ref], r)
		} else {
			ok++
		}
	}

	if len(brokenByRef) > 0 {
		fmt.Fprintln(w, "[+] Broken Links by Referrer")
		referrers := make([]string, 0, len(brokenByRef))
		for ref := range brokenByRef {
			referrers = append(referrers, ref)
		}
		slices.Sort(referrers)

		for _, ref := range referrers {
			fmt.Fprintf(w, "Referrer: %s\n", ref)

			for _, b := range brokenByRef[ref] {
				if b.Err != nil {
					fmt.Fprintf(w, "  -> %s (error: %v)\n", b.URL, b.Err)
				} else {
					fmt.Fprintf(w, "  -> %s (statusCode: %d)\n", b.URL, b.StatusCode)
				}
			}
			fmt.Fprintln(w)
		}
	}

	fmt.Fprintf(w, "Checked: %d | OK: %d | Broken: %d\n", total, ok, broken)
}

func writeVerbose(w io.Writer, results []crawl.Result) {
	fmt.Fprintln(w, "[+] Verbose Output")
	total := len(results)
	for i, r := range results {
		status := fmt.Sprintf("%d", r.StatusCode)
		if r.Err != nil {
			status = fmt.Sprintf("err (%v)", r.Err)
		}
		fmt.Fprintf(w, "(%d/%d) [Status: %s] [Depth: %d] -> %s\n", i+1, total, status, r.Depth, r.URL)
	}
	fmt.Fprintln(w)
}
