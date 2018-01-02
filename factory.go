package brink

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/brink/store"
	"golang.org/x/net/publicsuffix"
)

// NewCrawler returns an Crawler initialized with default values.
func NewCrawler(rootDomain string) (*Crawler, error) {
	rootDomainURL, err := schemeAndHost(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url %q: %v", rootDomain, err)
	}

	c := Crawler{
		RootDomain:       rootDomainURL,
		allowedDomains:   store.New(),
		visitedURLs:      store.New(),
		ignoredGETParams: store.New(),
		reqHeaders:       store.New(),
		handlers:         make(map[int]func(linkedFrom string, url string, status int, body string, cached bool)),
		client:           &http.Client{},
		opts: CrawlOptions{
			MaxContentLength: DefaultMaxContentLength,
			URLBufferSize:    100,
			WorkerCount:      10,
		},
	}

	c.urls = make(chan Link, c.opts.URLBufferSize)

	c.AllowDomains(rootDomainURL)

	return &c, nil
}

// NewCrawlerWithOpts returns a Crawler initialized with the provided CrawlOptions
// struct.
func NewCrawlerWithOpts(rootDomain string, userOptions CrawlOptions) (*Crawler, error) {
	c, err := NewCrawler(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed creating new crawler: %v", err)
	}

	// Headers
	if userOptions.Headers != nil {
		for k, v := range userOptions.Headers {
			c.reqHeaders.Store(k, v)
		}
	}

	// Domains
	err = setupDomains(&c.allowedDomains, c.RootDomain, userOptions.AllowedDomains)
	if err != nil {
		return nil, fmt.Errorf("allowed domains setup: %v", err)
	}

	// Cookies
	if userOptions.Cookies != nil {
		c.client.Jar, err = fillCookieJar(rootDomain, userOptions.Cookies)
		if err != nil {
			return nil, fmt.Errorf("cookie setup: %v", err)
		}
	}

	// Content length
	c.opts.MaxContentLength = getMaxContentLength(userOptions.MaxContentLength)

	// Ignore GET Parameters
	for _, v := range userOptions.IgnoreGETParameters {
		c.ignoredGETParams.StoreKey(v)
	}

	// Authentication
	err = configureAuth(c, userOptions.AuthType, userOptions.User, userOptions.Pass)
	if err != nil {
		return nil, fmt.Errorf("failed setting up auth: %v", err)
	}

	// Make sure we overwrite the default channel size in case it is specified in the
	// userOptions
	if userOptions.URLBufferSize != 0 {
		close(c.urls)
		c.urls = make(chan Link, userOptions.URLBufferSize)
	}

	if userOptions.WorkerCount > 0 {
		c.opts.WorkerCount = userOptions.WorkerCount
	}

	c.opts.FuzzyGETParameterChecks = userOptions.FuzzyGETParameterChecks

	return c, nil
}

// NewCrawlerFromToml reads up a file and parses it as a toml property file.
func NewCrawlerFromToml(filename string) (*Crawler, error) {
	var opts CrawlOptions

	if _, err := toml.DecodeFile(filename, &opts); err != nil {
		return nil, fmt.Errorf("failed decoding file: %v", err)
	}

	c, err := NewCrawlerWithOpts(opts.EntryPoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed creating crawler: %v", err)
	}

	return c, nil
}

func setupDomains(allowedDomains *store.CStore, rootDomain string, otherDomains []string) error {
	otherDomains = append(otherDomains, rootDomain)

	for _, domain := range otherDomains {
		url, err := schemeAndHost(domain)
		if err != nil {
			return fmt.Errorf("failed parsing allowed domain url %q: %v", domain, err)
		}

		allowedDomains.Store(url, "")
	}

	return nil
}

func fillCookieJar(rootDomain string, cookieMap map[string][]*http.Cookie) (http.CookieJar, error) {
	cj, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed creating cookie jar: %v", err)
	}

	urlCookieMap := make(map[string][]*http.Cookie)
	for _, cookies := range cookieMap {
		for _, cookie := range cookies {
			cks, ok := urlCookieMap[cookie.Domain]
			if !ok {
				cks = make([]*http.Cookie, 0)
			}

			cks = append(cks, cookie)

			urlCookieMap[cookie.Domain] = cks
		}
	}

	for u, cookies := range urlCookieMap {
		if u == "" {
			u = rootDomain
		}
		parsedURL, err := url.ParseRequestURI(u)
		if err != nil {
			return nil, fmt.Errorf("failed parsing url %q for cookie: %v", parsedURL, err)
		}

		cj.SetCookies(parsedURL, cookies)
	}

	return cj, nil
}

func getMaxContentLength(maxCL int64) int64 {
	switch maxCL {
	case 0:
		return DefaultMaxContentLength
	case -1:
		return UnlimitedMaxContentlength
	}

	return maxCL
}

func configureAuth(c *Crawler, authType int, user, pass string) error {
	switch authType {
	case AuthNone:
		return nil
	case AuthBasic:
		return configureBasicAuth(c, user, pass)
	}

	return nil
}

func configureBasicAuth(c *Crawler, user, pass string) error {
	userPass := fmt.Sprintf("%s:%s", user, pass)
	encodedUserPass := base64.StdEncoding.EncodeToString([]byte(userPass))

	c.reqHeaders.Store("Authorization", fmt.Sprintf("Basic %s", encodedUserPass))

	return nil
}
