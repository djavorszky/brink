package brink

import (
	"net/http"

	"github.com/djavorszky/brink/store"
)

// AuthType constants represent what type of authentication to use
// when visiting the pages
const (
	AuthNone = iota
	AuthBasic
)

// ContentLengths are used when fetching a page. If the content length
// as reported by the server is lar ger than is specified, the page won't
// be downloaded.
const (
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

	// urls is the channel that goroutines will send URLs to and
	// receive URLs from when visiting and processing pages.
	urls chan Link

	reqHeaders       store.CStore
	allowedDomains   store.CStore
	visitedURLs      store.CStore
	ignoredGETParams store.CStore
}

// CrawlOptions contains options for the crawler
type CrawlOptions struct {
	AuthType int    `toml:"auth-type"`
	User     string `toml:"user"`
	Pass     string `toml:"pass"`

	// URLBufferSize is the amount of URLs that can be waiting to be visited.
	URLBufferSize int `toml:"url-buffer-size"`

	// WorkerCount specifies the number of goroutines that will work on crawling the domains.
	WorkerCount int `toml:"worker-count"`

	// MaxContentLength specifies the maximum size of pages to be crawled. Setting it to 0
	// will default to 512Kb. Set it to -1 to allow unlimited size
	MaxContentLength int64 `toml:"max-content-length"`

	// AllowedDomains will be used to check whether a domain is allowed to be crawled or not.
	AllowedDomains []string `toml:"allowed-domains"`

	// Cookies holds a mapping for URLs -> list of cookies to be added to all requests
	Cookies map[string][]*http.Cookie `toml:"cookies"`

	// Headers holds a mapping for key->values to be added to all requests
	Headers map[string]string `toml:"headers"`

	// Ignore certain GET parameters when comparing whether an URL has been visited or not
	IgnoreGETParameters []string `toml:"ignore-get-parameters"`

	// FuzzyGETParameterChecks will decide whether to try to do exact matches for parameters.
	// If set to false, GET parameters are only ignored if they are an exact match. If set
	// to true, they are checked with a substring fashion.
	FuzzyGETParameterChecks bool `toml:"fuzzy-get-parameter-checks"`

	// todo: add ctx
	// todo: add proxy support
	// todo: add beforeFunc and afterFunc
}
