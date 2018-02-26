package brink

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/brink/store"
)

const (
	defaultMaxContentLength      = 512 * 1024 // 512Kb
	defaultURLBufferSize         = 10000
	defaultWorkerCount           = 10
	defaultIdleWorkCheckInterval = 5000

	unlimitedMaxContentlength = math.MaxInt64 // 4,61 exabytes

	authorizationHeaderName = "Authorization"
)

// NewCrawler returns an Crawler initialized with default values.
func NewCrawler(rootDomain string) (*Crawler, error) {
	scheme, host, err := schemeAndHost(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url %q: %v", rootDomain, err)
	}

	rootDomainURL := fmt.Sprintf("%s://%s", scheme, host)

	c := Crawler{
		RootDomain:       rootDomainURL,
		allowedDomains:   store.New(),
		visitedURLs:      store.New(),
		ignoredGETParams: store.New(),
		reqHeaders:       store.New(),
		forbiddenPaths:   store.New(),
		handlers:         make(map[int]func(linkedFrom string, url string, status int, body string, cached bool)),
		client:           &http.Client{},
		opts: CrawlOptions{
			MaxContentLength:      defaultMaxContentLength,
			URLBufferSize:         defaultURLBufferSize,
			WorkerCount:           defaultWorkerCount,
			IdleWorkCheckInterval: defaultIdleWorkCheckInterval,
			Cookies:               make(map[string]*http.Cookie),
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
		for _, cookie := range userOptions.Cookies {
			c.opts.Cookies[cookie.Name] = cookie
		}
	}

	// Session cookie names
	if userOptions.SessionCookieNames != nil {
		c.opts.SessionCookieNames = userOptions.SessionCookieNames
	}

	// Content length
	c.opts.MaxContentLength = getMaxContentLength(userOptions.MaxContentLength)

	// Idle check interval
	if userOptions.IdleWorkCheckInterval > 0 {
		c.opts.IdleWorkCheckInterval = userOptions.IdleWorkCheckInterval
	}

	// Ignore GET Parameters
	for _, v := range userOptions.IgnoreGETParameters {
		c.ignoredGETParams.StoreKey(v)
	}

	// Forbidden paths
	for _, v := range userOptions.ForbiddenPaths {
		c.forbiddenPaths.StoreKey(v)
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
		scheme, host, err := schemeAndHost(domain)
		if err != nil {
			return fmt.Errorf("failed parsing allowed domain url %q: %v", domain, err)
		}

		allowedDomains.Store(fmt.Sprintf("%s://%s", scheme, host), "")
	}

	return nil
}

func getMaxContentLength(maxCL int64) int64 {
	switch maxCL {
	case 0:
		return defaultMaxContentLength
	case -1:
		return unlimitedMaxContentlength
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

	c.reqHeaders.Store(authorizationHeaderName, fmt.Sprintf("Basic %s", encodedUserPass))

	return nil
}
