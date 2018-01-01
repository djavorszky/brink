package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/brink"
)

type userOpts struct {
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

	// Entrypoint is the first url that will be fetched.
	EntryPoint string `toml:"entrypoint"`

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
}

func main() {
	config := flag.String("conf", "brink.toml", "Specify the configuration filename to be used")

	flag.Parse()

	var opts userOpts
	if _, err := toml.DecodeFile(*config, &opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := brink.NewCrawlerWithOpts(opts.EntryPoint,
		brink.CrawlOptions{
			AuthType:                opts.WorkerCount,
			User:                    opts.User,
			Pass:                    opts.Pass,
			URLBufferSize:           opts.URLBufferSize,
			WorkerCount:             opts.WorkerCount,
			MaxContentLength:        opts.MaxContentLength,
			AllowedDomains:          opts.AllowedDomains,
			Cookies:                 opts.Cookies,
			Headers:                 opts.Headers,
			IgnoreGETParameters:     opts.IgnoreGETParameters,
			FuzzyGETParameterChecks: opts.FuzzyGETParameterChecks,
		},
	)
	if err != nil {
		fmt.Printf("Failed initializing crawler: %v\n", err)
		os.Exit(1)
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		c.Stop()
	}()

	c.HandleDefaultFunc(handler)
	c.HandleFunc(http.StatusNotFound, notFoundHandler)

	c.Start()
}

func handler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		log.Printf("CACHED: %s -> %s: %d ", linkedFrom, url, status)
	} else {
		log.Printf("%s -> %s: %d ", linkedFrom, url, status)
	}
}

func notFoundHandler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		log.Printf("CACHED: 404: %s", url)
	} else {
		log.Printf("404: %s", url)
	}
}
