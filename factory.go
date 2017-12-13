package brink

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

// NewCrawler returns an Crawler initialized with default values.
func NewCrawler(rootDomain string) (*Crawler, error) {
	rootDomainURL, err := schemeAndHost(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url %q: %v", rootDomain, err)
	}

	c := Crawler{
		RootDomain:     rootDomainURL,
		allowedDomains: make(map[string]bool),
		visitedURLs:    make(map[string]bool),
		handlers:       make(map[int]func(url string, status int, body string)),
		client:         &http.Client{},
		opts:           CrawlOptions{MaxContentLength: DefaultMaxContentLength},
	}

	c.allowedDomains[rootDomainURL] = true

	return &c, nil
}

// NewCrawlerWithOpts returns a Crawler initialized with the provided CrawlOptions
// struct.
func NewCrawlerWithOpts(rootDomain string, userOptions CrawlOptions) (*Crawler, error) {
	rootDomainURL, err := schemeAndHost(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url %q: %v", rootDomain, err)
	}

	c := Crawler{
		RootDomain:     rootDomainURL,
		allowedDomains: make(map[string]bool),
		visitedURLs:    make(map[string]bool),
		handlers:       make(map[int]func(url string, status int, body string)),
		client:         &http.Client{},
		opts:           userOptions,
	}

	c.allowedDomains[rootDomainURL] = true
	for _, domain := range userOptions.AllowedDomains {
		url, err := schemeAndHost(domain)
		if err != nil {
			return nil, fmt.Errorf("failed parsing allowed domain url %q: %v", domain, err)
		}

		c.allowedDomains[url] = true
	}

	if userOptions.Cookies != nil {
		c.client.Jar, err = fillCookieJar(userOptions.Cookies)
		if err != nil {
			return nil, fmt.Errorf("cookie setup: %v", err)
		}
	}

	switch userOptions.MaxContentLength {
	case 0:
		c.opts.MaxContentLength = DefaultMaxContentLength
	case -1:
		c.opts.MaxContentLength = UnlimitedMaxContentlength
	}

	return &c, nil
}

func fillCookieJar(cookieMap map[string][]*http.Cookie) (http.CookieJar, error) {
	cj, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed creating cookie jar: %v", err)
	}

	for u, cookies := range cookieMap {
		parsedURL, err := url.ParseRequestURI(u)
		if err != nil {
			return nil, fmt.Errorf("failed parsing url %q for cookie: %v", parsedURL, err)
		}

		cj.SetCookies(parsedURL, cookies)
	}

	return cj, nil
}
