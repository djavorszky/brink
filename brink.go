package brink

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Start starts the crawler at the specified rootDomain. It will scrape the page for
// links and then visit each of them, provided the domains are allowed. It will keep
// repeating this process on each page until it runs out of pages to visit.
//
// Start requires at least one handler to be registered, otherwise errors out.
func (c *Crawler) Start() error {
	// Prefetch checks
	if c.RootDomain == "" {
		return fmt.Errorf("root domain not specified")
	}

	if c.defaultHandler == nil && len(c.handlers) == 0 {
		return fmt.Errorf("no handlers specified")
	}

	// Spawn workers
	var wg sync.WaitGroup
	c.spawnWorkers(&wg)

	c.urls <- Link{LinkedFrom: c.RootDomain, Href: c.RootDomain}

	// Spawn checker
	go func() {
		interval := time.Duration(c.opts.IdleWorkCheckInterval)

	ticker:
		for range time.Tick(interval * time.Millisecond) {
			select {
			case url := <-c.urls:
				if !c.stopping {
					c.urls <- url
				}
			default:
				log.Println("No urls to parse, exiting.")
				c.Stop()
				break ticker
			}
		}
	}()

	wg.Wait()

	return nil
}

func (c *Crawler) spawnWorkers(wg *sync.WaitGroup) {
	wg.Add(c.opts.WorkerCount)

	for i := 0; i < c.opts.WorkerCount; i++ {
		name := fmt.Sprintf("worker-%d", i+1)
		log.Printf("Spawning %s", name)

		go func(name string) {
			defer wg.Done()

		loop:
			for link := range c.urls {
				url, err := c.normalizeURL(link.Href)
				if err != nil {
					// Debug..
					log.Printf("%s: failed normalize: %v", name, err)
					continue
				}

				if st, ok := c.visitedURLs.Load(url); ok {
					st, _ := strconv.Atoi(st)
					if f, ok := c.handlers[st]; ok {
						f(link.LinkedFrom, url, st, "", true)
					} else {
						c.defaultHandler(link.LinkedFrom, url, st, "", true)
					}

					continue
				}

				st, bod, err := c.Fetch(url)
				if err != nil {
					// Debug..
					//log.Printf("%s: failed fetch: %v", name, err)
					continue
				}

				c.visitedURLs.Store(url, strconv.Itoa(st))

				if f, ok := c.handlers[st]; ok {
					f(link.LinkedFrom, url, st, "", false)
				} else {
					c.defaultHandler(link.LinkedFrom, url, st, "", false)
				}

				if st != http.StatusOK {
					continue
				}

				// Parse links and send them all to the urls channel
				links, err := AbsoluteLinksIn(link.Href, link.Href, bod, true)
				if err != nil {
					log.Printf("err in AbsLinksIn: %v", err)
					continue
				}

				for _, l := range links {
					if l.Href == "" {
						continue
					}

					if c.stopping {
						break loop
					}

					c.urls <- l
				}

				//log.Printf("%s: count: %d, linkCount: %d", name, count, lc)
			}
		}(name)
	}
}

// Stop attempts to stop the crawler.
func (c *Crawler) Stop() {
	log.Println("Received signal to stop... Will finish cached runs.")
	c.stopping = true
	close(c.urls)
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

	// Add headers
	if c.reqHeaders.Size() != 0 {
		for key, value := range c.reqHeaders.ToMap() {
			req.Header.Add(key, value)
		}
	}

	// Add cookies
	reqCookies := c.cookies()
	if len(reqCookies) != 0 {
		for _, cookie := range reqCookies {
			log.Printf("Cookie: %v", cookie)
			req.AddCookie(cookie)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("get failed: %v", err)
	}
	defer resp.Body.Close()

	// Add response cookies
	respCookies := resp.Cookies()
	if len(respCookies) != 0 {

	}

	scheme, host, err := schemeAndHost(url)
	if err != nil {
		return 0, nil, fmt.Errorf("malformed url: %v", err)
	}

	domain := fmt.Sprintf("%s://%s", scheme, host)
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
func (c *Crawler) HandleDefaultFunc(h func(linkedFrom string, url string, status int, body string, cached bool)) {
	c.defaultHandler = h
}

// HandleFunc is used to register a function to be called when a new page is
// found with the specified status. Subsequent calls to register functions
// to the same statuses will silently overwrite previously set handlers, if any.
func (c *Crawler) HandleFunc(status int, h func(linkedFrom string, url string, status int, body string, cached bool)) {
	c.handlers[status] = h
}

func (c *Crawler) seenURL(url string) bool {
	return c.visitedURLs.Contains(url)
}

func (c *Crawler) domainAllowed(domain string) bool {
	_, ok := c.allowedDomains.Load(domain)

	return ok
}

func (c *Crawler) cookies() (cks []*http.Cookie) {
	c.opts.cmu.RLock()
	defer c.opts.cmu.RUnlock()
	copy(cks, c.opts.Cookies)

	return cks
}
