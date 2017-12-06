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
)

// Walker represents a web crawler, starting from a RootDomain
// and visiting all the links in the AllowedDomains map. It will only
// download the body of an URL if it is less than MaxContentLength.
type Walker struct {
	RootDomain string
	AuthType   int
	User, Pass string

	MaxContentLength int64
	//	WorkerCount    int

	client *http.Client

	// dmu is the RWMutext for the allowed domain
	dmu            *sync.RWMutex
	allowedDomains map[string]bool

	// vmu is the RWMutex for the visits
	vmu    *sync.RWMutex
	visits map[string]bool

	// ctx...?
	// ctx context.Context

	// Todo: add proxy

	// Handlers...
	defaultHandler func(url, body string, status int)
	handlers       map[int]func(url, body string, status int)
}

// NewWalker returns an initialized Walker.
func NewWalker(rootDomain string) (Walker, error) {
	url, err := schemeAndHost(rootDomain)
	if err != nil {
		return Walker{}, fmt.Errorf("failed parsing url %q: %v", rootDomain, err)
	}

	w := Walker{
		RootDomain: url,
	}

	w.allowedDomains = make(map[string]bool)
	w.visits = make(map[string]bool)
	w.handlers = make(map[int]func(url, body string, status int))

	w.allowedDomains[w.RootDomain] = true

	w.MaxContentLength = 512 * 1024 // 512Kbyte

	return w, nil
}

// Start starts the walker.
func (w Walker) Start() error {
	if w.RootDomain == "" {
		return fmt.Errorf("root domain not specified")
	}

	if w.defaultHandler == nil && len(w.handlers) == 0 {
		return fmt.Errorf("no handlers specified")
	}

	return nil
}

// Stop attempts to stop the walker.
func (w Walker) Stop() error {
	return nil
}

// AllowDomains instructs the walker which domains it is allowed
// to visit. The RootDomain is automatically added to this list.
// Domains not allowed will be checked for http status, but will
// not be traversed.
//
// Subsequent calls to AllowDomains adds to the list of domains
// allowed to the walker to traverse.
func (w Walker) AllowDomains(domains ...string) {
	w.dmu.Lock()
	defer w.dmu.Unlock()

	for _, domain := range domains {
		w.allowedDomains[domain] = true
	}
}

// HandleDefaultFunc will be called for all pages returned by a status
// which doesn't have a seperate handler defined by HandleFunc. Subsequent
// calls to HandleDefaultFunc will overwrite the previously set handlers,
// if any.
func (w Walker) HandleDefaultFunc(h func(url, body string, status int)) {
	w.defaultHandler = h
}

// HandleFunc is used to register a function to be called when a new page is
// found with the specified status. Subsequent calls to register functions
// to the same statuses will silently overwrite previously set handlers, if any.
func (w Walker) HandleFunc(status int, h func(url, body string, status int)) {
	w.handlers[status] = h
}

func (w Walker) seen(url string) bool {
	w.vmu.RLock()
	defer w.vmu.RUnlock()

	_, seen := w.visits[url]

	return seen
}

func (w Walker) saveVisit(url string) {
	w.vmu.Lock()
	defer w.vmu.Unlock()

	w.visits[url] = true
}

func (w Walker) allowed(url string) bool {
	w.dmu.RLock()
	defer w.dmu.RUnlock()

	_, ok := w.allowedDomains[url]

	return ok
}

func (w Walker) fetch(url string) (status int, body string, err error) {
	if w.seen(url) {
		return 0, "", nil
	}

	resp, err := w.client.Get(url)
	if err != nil {
		return 0, "", fmt.Errorf("get failed: %v", err)
	}
	defer resp.Body.Close()

	// if URL is not allowed, return with only its status code
	if !w.allowed(url) {
		return resp.StatusCode, "", nil
	}

	// if response size is too large (or unknown), return early with
	// only the status code
	if resp.ContentLength > w.MaxContentLength {
		return resp.StatusCode, "", nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed reading response body: %v", err)
	}

	return resp.StatusCode, string(b), nil
}
