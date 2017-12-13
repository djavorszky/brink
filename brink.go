package brink

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// AuthType constants represent what type of authentication to use
// when visiting the pages
const (
	AuthNone = iota
	AuthBasic

	DefaultMaxContentLength   = 512 * 1024          // 512Kb
	UnlimitedMaxContentlength = 9223372036854775807 // 9.22 exabytes
)

// Crawler represents a web crawler, starting from a RootDomain
// and visiting all the links in the AllowedDomains map. It will only
// download the body of an URL if it is less than MaxContentLength.
type Crawler struct {
	RootDomain string
	client     *http.Client
	opts       CrawlOptions

	// Handlers...
	defaultHandler func(url string, status int, body string)
	handlers       map[int]func(url string, status int, body string)

	// dmu is the RWMutext for the allowed domains
	dmu            sync.RWMutex
	allowedDomains map[string]bool

	// vmu is the RWMutex for the URLs already visited
	vmu         sync.RWMutex
	visitedURLs map[string]bool
}

// CrawlOptions contains options for the crawler
type CrawlOptions struct {
	AuthType   int
	User, Pass string

	// MaxContentLength specifies the maximum size of pages to be crawled. Setting it to 0
	// will default to 512Kb. Set it to -1 to allow unlimited size
	MaxContentLength int64

	// AllowedDomains will be used to check whether a domain is allowed to be crawled or not.
	AllowedDomains []string

	// Cookies holds a mapping for URLs -> list of cookies to be added to all requests
	Cookies map[string][]*http.Cookie

	// Headers holds a mapping for key->values to be added to all requests
	Headers map[string]string

	// todo: add auth
	// todo: add ctx
	// todo: add proxy support
	// todo: add beforeFunc and afterFunc
	// todo: add multiple workers
	// todo: only differentiate btw pages if their GET parameters actually differ (i.e. ignore order)
}

// Start starts the crawler at the specified rootDomain. It will scrape the page for
// links and then visit each of them, provided the domains are allowed. It will keep
// repeating this process on each page until it runs out of pages to visit.
//
// Start requires at least one handler to be registered, otherwise errors out.
func (c *Crawler) Start() error {
	if c.RootDomain == "" {
		return fmt.Errorf("root domain not specified")
	}

	if c.defaultHandler == nil && len(c.handlers) == 0 {
		return fmt.Errorf("no handlers specified")
	}

	return nil
}

// Stop attempts to stop the crawler.
func (c *Crawler) Stop() error {
	return nil
}

// AllowDomains instructs the crawler which domains it is allowed
// to visit. The RootDomain is automatically added to this list.
// Domains not allowed will be checked for http status, but will
// not be traversed.
//
// Subsequent calls to AllowDomains adds to the list of domains
// allowed to the crawler to traverse.
func (c *Crawler) AllowDomains(domains ...string) {
	c.dmu.Lock()
	defer c.dmu.Unlock()

	for _, domain := range domains {
		c.allowedDomains[domain] = true
	}
}

// Fetch fetches the URL and returns its status, body and/or any errors it
// encountered.
func (c *Crawler) Fetch(url string) (status int, body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("failed creating new request: %v", err)
	}

	if c.opts.Headers != nil {
		for key, value := range c.opts.Headers {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("get failed: %v", err)
	}
	defer resp.Body.Close()

	// if URL is not allowed, return with only its status code
	if !c.urlAllowed(url) {
		return resp.StatusCode, nil, nil
	}

	// if response size is too large (or unknown), return early with
	// only the status code
	if resp.ContentLength > c.opts.MaxContentLength {
		return resp.StatusCode, nil, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed reading response body: %v", err)
	}

	return resp.StatusCode, b, nil
}

// HandleDefaultFunc will be called for all pages returned by a status
// which doesn't have a seperate handler defined by HandleFunc. Subsequent
// calls to HandleDefaultFunc will overwrite the previously set handlers,
// if any.
func (c *Crawler) HandleDefaultFunc(h func(url string, status int, body string)) {
	c.defaultHandler = h
}

// HandleFunc is used to register a function to be called when a new page is
// found with the specified status. Subsequent calls to register functions
// to the same statuses will silently overwrite previously set handlers, if any.
func (c *Crawler) HandleFunc(status int, h func(url string, status int, body string)) {
	c.handlers[status] = h
}

func (c *Crawler) seenURL(url string) bool {
	c.vmu.RLock()
	defer c.vmu.RUnlock()

	_, seen := c.visitedURLs[url]

	return seen
}

func (c *Crawler) saveVisit(url string) {
	c.vmu.Lock()
	defer c.vmu.Unlock()

	c.visitedURLs[url] = true
}

func (c *Crawler) urlAllowed(url string) bool {
	c.dmu.RLock()
	defer c.dmu.RUnlock()

	_, ok := c.allowedDomains[url]

	return ok
}
