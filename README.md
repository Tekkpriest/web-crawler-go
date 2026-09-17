# Web Crawler for Broken Links

This is a small Web Crawler i built using Go.
Currently on the V1 iteration so to speak, but the next iteration will feature concurrent worker pools instead of the sequential flow available now.

With this project you can recursively check any domain for working and broken links up until a depth of n. 
It also will be documented in either text format using StdOut or formatted as a json.

Broken links won't be treated as an error but instead are being documented as well.

The program will exit with 3 possible status codes:

- 0 - all links good
- 1 - broken links found
- 2 - user errors (like missing url or flags)

This also allows for the project to be used in CI, if you so choose.


## Installation

```bash
git clone https://github.com/tekkpriest/web-crawler-go.git
cd web-crawler-go
go build -o crawler
```
This will clone the repo, change into the directory and build a binary called crawler in that directory.


## Usage

Open up your terminal and run 

```bash
./crawler -url "https://example.com"
```

with following flags available

`-url "https://example.com"`
hard requirement: enter http or https domain 

`-depth n`
check domain until depth of n, default is 2

`-timeout n`
set http request timeout (e.g. 5s, 500ms), default is 10s

`-json`
enable JSON output instead of text reports for piping into other tools/scripts, disabled by default

`-v`
enable verbose output, disabled by default


## Known Issues & Limitations

If a page's Content-Type header can't be parsed or its HTML fails to parse, the page itself still shows up in results 
but outgoing links are not followed (no error/log surfaced).

normalise() only lowers the scheme not the host (e.g. EXAMPLE.com/page and example.com/page count as two different pages), so
double results might be possible.

-json currently might write {} instead of a readable output when network errors are present.

Also only `<a href>` is currently being crawled, `<link>` tags are not present currently.

## Roadmap to next version

As mentioned above, concurrent worker pools are in the works, also robots.txt and rate limiting (delaying the crawler) are coming.
Then i want to make redirect chains more visible as well as enable csv export with -include / -exclude, as well as include HEAD with GET Fallbacks.
