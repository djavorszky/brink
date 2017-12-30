package brink

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

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

	st, bod, err := c.Fetch(c.RootDomain)
	if err != nil {
		return fmt.Errorf("failed initial fetch: %v", err)
	}

	links := LinksIn(bod, true)
	for _, l := range links {
		fmt.Println(l)
	}

	if f, ok := c.handlers[st]; ok {
		f(c.RootDomain, st, string(bod))
	} else {
		c.defaultHandler(c.RootDomain, st, string(bod))
	}

	return nil
}

// Stop attempts to stop the crawler.
func (c *Crawler) Stop() error {
	// todo: implement

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
	for _, domain := range domains {
		c.allowedDomains.StoreKey(domain)
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

	domain, err := schemeAndHost(url)
	if err != nil {
		return 0, nil, fmt.Errorf("malformed url: %v", err)
	}

	// if URL is not allowed, return with only its status code
	if !c.domainAllowed(domain) {
		return resp.StatusCode, nil, NotAllowed{domain}
	}

	// if response size is too large (or unknown), return early with
	// only the status code
	if resp.ContentLength > c.opts.MaxContentLength {
		return resp.StatusCode, nil, ContentTooLarge{url}
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
	normalizedURL, _ := c.normalizeURL(url)

	return c.visitedURLs.Contains(normalizedURL)
}

func (c *Crawler) saveVisit(url string) {
	normalizedURL, _ := c.normalizeURL(url)

	c.visitedURLs.StoreKey(normalizedURL)
}

func (c *Crawler) domainAllowed(domain string) bool {
	_, ok := c.allowedDomains.Load(domain)

	return ok
}
