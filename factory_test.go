package brink

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewCrawler(t *testing.T) {
	type args struct {
		rootDomain string
	}
	tests := []struct {
		name           string
		args           args
		wantRootDomain string
		wantErr        bool
	}{
		{"Missing schema", args{"google.com"}, "", true},
		{"URL", args{"https://liferay.com/"}, "https://liferay.com", false},
		{"URL with path", args{"https://liferay.com/web/guest/home"}, "https://liferay.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawler(tt.args.rootDomain)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCrawler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.RootDomain != tt.wantRootDomain {
				t.Errorf("rootDomain mismatch: %s vs %s", got.RootDomain, tt.args.rootDomain)
			}

			// check defaults
			if got.opts.MaxContentLength != defaultMaxContentLength {
				t.Errorf("MaxContentLength mismatch: %d vs %d", got.opts.MaxContentLength, defaultMaxContentLength)
			}

			if got.opts.URLBufferSize != defaultURLBufferSize {
				t.Errorf("URLBufferSize mismatch: %d vs %d", got.opts.URLBufferSize, defaultURLBufferSize)
			}

			if got.opts.WorkerCount != defaultWorkerCount {
				t.Errorf("WorkerCount mismatch: %d vs %d", got.opts.WorkerCount, defaultWorkerCount)
			}

			if got.opts.IdleWorkCheckInterval != defaultIdleWorkCheckInterval {
				t.Errorf("IdleWorkCheckInterval mismatch: %d vs %d", got.opts.IdleWorkCheckInterval, defaultIdleWorkCheckInterval)
			}

			// Check initializations. If below succeed, all is good.
			got.allowedDomains.Store("testKey", "testValue")
			got.visitedURLs.Store("testKey", "testValue")
			got.ignoredGETParams.Store("testKey", "testValue")
			got.reqHeaders.Store("testKey", "testValue")
			got.handlers[200] = func(linkedFrom, url string, st int, bod string, cached bool) {}
			got.urls <- Link{}

			if !got.allowedDomains.Contains(tt.wantRootDomain) {
				t.Errorf("rootDomain was not allowed.")
			}
		})
	}
}

func TestNewCrawlerWithOpts(t *testing.T) {
	defaultCrawler, _ := NewCrawler("https://www.liferay.com")
	// defaultOpts := CrawlOptions{
	// 	MaxContentLength:      defaultMaxContentLength,
	// 	URLBufferSize:         defaultURLBufferSize,
	// 	WorkerCount:           defaultWorkerCount,
	// 	IdleWorkCheckInterval: defaultIdleWorkCheckInterval,
	// }

	type args struct {
		rootDomain  string
		userOptions CrawlOptions
	}
	tests := []struct {
		name    string
		args    args
		want    *Crawler
		wantErr bool
	}{
		{"emptyOpts", args{defaultCrawler.RootDomain, CrawlOptions{}}, defaultCrawler, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawlerWithOpts(tt.args.rootDomain, tt.args.userOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCrawlerWithOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = compareCrawlers(got, tt.want); err != nil {
				t.Errorf("Crawler mismatch: %v", err)
			}
		})
	}
}

func Test_getMaxContentLength(t *testing.T) {
	type args struct {
		maxCL int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"Default", args{0}, defaultMaxContentLength},
		{"Unlimited", args{-1}, unlimitedMaxContentlength},
		{"Some value", args{512000}, 512000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMaxContentLength(tt.args.maxCL); got != tt.want {
				t.Errorf("getMaxContentLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCrawlerFromToml(t *testing.T) {
	contents := []byte(`auth-type = 1
user = "testUser"
pass = "testPassword"
url-buffer-size = 5000
worker-count = 2
max-content-length = 10000
entrypoint = "http://example.com"
allowed-domains = ["http://www.example.com"]
ignore-get-parameters = ["redirect"]
fuzzy-get-parameter-checks = true
idle-work-check-interval = 2000
[[cookies]]
Name = "Cookie Name"
Value = "Cookie Value"
Path = "/"
Domain = "http://example.com"
Expires = 2018-12-31T22:59:59Z
RawExpires = ""
MaxAge = 0
Secure = true
HttpOnly = false
Raw = ""

[[cookies]]
Name = "Second Cookie Name"
Value = "Second Cookie Value"
Path = "/"
Domain = "http://example.com"
Expires = 2018-12-31T22:59:59Z
RawExpires = ""
MaxAge = 0
Secure = true
HttpOnly = false
Raw = ""
[headers]
header-name = "header-value"`)

	f, err := ioutil.TempFile(".", "test")
	if err != nil {
		t.Errorf("Failed creating temp file: %v", err)
		return
	}
	defer f.Close()
	defer os.Remove(f.Name())

	err = ioutil.WriteFile(f.Name(), contents, os.ModePerm)
	if err != nil {
		t.Errorf("Failed writing temp file: %v", err)
		return
	}
	date, _ := time.Parse(time.RFC3339, "2018-12-31T22:59:59Z")

	opts := CrawlOptions{
		User:                    "testUser",
		Pass:                    "testPassword",
		URLBufferSize:           5000,
		WorkerCount:             2,
		MaxContentLength:        10000,
		EntryPoint:              "http://example.com",
		AllowedDomains:          []string{"http://www.example.com"},
		IgnoreGETParameters:     []string{"redirect"},
		FuzzyGETParameterChecks: true,
		IdleWorkCheckInterval:   2000,
		Cookies: map[string]*http.Cookie{
			"CookieName": &http.Cookie{
				Domain:  "http://example.com",
				Name:    "CookieName",
				Value:   "Cookie Value",
				Path:    "/",
				Expires: date,
				Secure:  true,
			},
			"SecondCookieName": &http.Cookie{
				Domain:  "http://example.com",
				Name:    "SecondCookieName",
				Value:   "Second Cookie Value",
				Path:    "/",
				Expires: date,
				Secure:  true,
			},
		},
		Headers: map[string]string{"header-name": "header-value"},
	}

	c, _ := NewCrawlerWithOpts(opts.EntryPoint, opts)

	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    *Crawler
		wantErr bool
	}{
		{"overwrite", args{f.Name()}, c, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawlerFromToml(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCrawlerFromToml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err = compareCrawlers(got, tt.want); err != nil {
				t.Errorf("Mismatching crawlers: %v", err)
			}
		})
	}
}

func compareCrawlers(got, want *Crawler) error {
	if got.RootDomain != want.RootDomain {
		return fmt.Errorf("NewCrawlerFromToml() RootDomain mismatch: %s vs %s", got.RootDomain, want.RootDomain)
	}

	if got.opts.EntryPoint != want.opts.EntryPoint {
		return fmt.Errorf("NewCrawlerFromToml() EntryPoint mismatch: %s vs %s", got.opts.EntryPoint, want.opts.EntryPoint)
	}

	if got.opts.AuthType != want.opts.AuthType {
		return fmt.Errorf("NewCrawlerFromToml() AuthType mismatch: %d vs %d", got.opts.AuthType, want.opts.AuthType)
	}

	if got.opts.User != want.opts.User {
		return fmt.Errorf("NewCrawlerFromToml() User mismatch: %s vs %s", got.opts.User, want.opts.User)
	}

	if got.opts.Pass != want.opts.Pass {
		return fmt.Errorf("NewCrawlerFromToml() Pass mismatch: %s vs %s", got.opts.Pass, want.opts.Pass)
	}

	if got.opts.URLBufferSize != want.opts.URLBufferSize {
		return fmt.Errorf("NewCrawlerFromToml() URLBufferSize mismatch: %d vs %d", got.opts.URLBufferSize, want.opts.URLBufferSize)
	}

	if got.opts.WorkerCount != want.opts.WorkerCount {
		return fmt.Errorf("NewCrawlerFromToml() WorkerCount mismatch: %d vs %d", got.opts.WorkerCount, want.opts.WorkerCount)
	}

	if got.opts.MaxContentLength != want.opts.MaxContentLength {
		return fmt.Errorf("NewCrawlerFromToml() MaxContentLength mismatch: %d vs %d", got.opts.MaxContentLength, want.opts.MaxContentLength)
	}

	if !reflect.DeepEqual(got.opts.AllowedDomains, want.opts.AllowedDomains) {
		return fmt.Errorf("NewCrawlerFromToml() AllowedDomains mismatch: %s vs %s", got.opts.AllowedDomains, want.opts.AllowedDomains)
	}

	if !reflect.DeepEqual(got.opts.Headers, want.opts.Headers) {
		return fmt.Errorf("NewCrawlerFromToml() Headers mismatch: %s vs %s", got.opts.Headers, want.opts.Headers)
	}

	for i, c := range got.opts.Cookies {
		if err := compareCookies(c, want.opts.Cookies[i]); err != nil {
			return fmt.Errorf("NewCrawlerFromToml() Cookies mismatch: %v", err)
		}
	}

	if !reflect.DeepEqual(got.opts.IgnoreGETParameters, want.opts.IgnoreGETParameters) {
		return fmt.Errorf("NewCrawlerFromToml() IgnoreGETParameters mismatch: %s vs %s", got.opts.IgnoreGETParameters, want.opts.IgnoreGETParameters)
	}

	if got.opts.FuzzyGETParameterChecks != want.opts.FuzzyGETParameterChecks {
		return fmt.Errorf("NewCrawlerFromToml() FuzzyGETParameterChecks mismatch: %t vs %t", got.opts.FuzzyGETParameterChecks, want.opts.FuzzyGETParameterChecks)
	}

	if got.opts.IdleWorkCheckInterval != want.opts.IdleWorkCheckInterval {
		return fmt.Errorf("NewCrawlerFromToml() IdleWorkCheckInterval mismatch: %d vs %d", got.opts.IdleWorkCheckInterval, want.opts.IdleWorkCheckInterval)
	}

	return nil
}

func compareCookies(c1, c2 *http.Cookie) error {
	if c1.Domain != c2.Domain {
		return fmt.Errorf("Domain mismatch: %v vs %v", c1.Domain, c2.Domain)
	}

	if c1.Name != c2.Name {
		return fmt.Errorf("Name mismatch: %v vs %v", c1.Name, c2.Name)
	}

	if c1.Value != c2.Value {
		return fmt.Errorf("Value mismatch: %v vs %v", c1.Value, c2.Value)
	}

	if c1.Expires != c2.Expires {
		return fmt.Errorf("Expires mismatch: %v vs %v", c1.Expires, c2.Expires)
	}

	if c1.Secure != c2.Secure {
		return fmt.Errorf("Secure mismatch: %v vs %v", c1.Secure, c2.Secure)
	}

	if c1.String() != c2.String() {
		return fmt.Errorf("String() mismatch: %v vs %v", c1.String(), c2.String())
	}

	return nil
}
