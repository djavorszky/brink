package brink

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/publicsuffix"
)

/*
func TestNewCrawler(t *testing.T) {
	var tClient *http.Client

	testCrawler := &Crawler{
		RootDomain:     "https://liferay.com",
		allowedDomains: store.New(),
		visitedURLs:    store.New(),
		handlers:       make(map[int]func(url string, status int, body string)),
		client:         tClient,
		opts:           CrawlOptions{MaxContentLength: DefaultMaxContentLength},
	}

	testCrawler.AllowDomains("https://liferay.com")

	type args struct {
		rootDomain string
	}
	tests := []struct {
		name    string
		args    args
		want    *Crawler
		wantErr bool
	}{
		{"Missing schema", args{"google.com"}, nil, true},
		{"URL", args{"https://liferay.com/"}, testCrawler, false},
		{"URL with path", args{"https://liferay.com/web/guest/home"}, testCrawler, false},
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
			got.client = tClient
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCrawler() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
func TestNewCrawlerWithOpts(t *testing.T) {
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
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCrawlerWithOpts(tt.args.rootDomain, tt.args.userOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCrawlerWithOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCrawlerWithOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fillCookieJar(t *testing.T) {
	type args struct {
		cookieMap map[string][]*http.Cookie
	}

	refNoCookies, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	refTwoCookies, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	refDiffDomainCookies, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	c1 := &http.Cookie{Domain: "https://liferay.com", Name: "User", Value: "Test"}
	c2 := &http.Cookie{Domain: "https://liferay.com", Name: "Options", Value: "No"}

	d1, _ := url.ParseRequestURI("https://liferay.com")
	d2, _ := url.ParseRequestURI("https://dev.liferay.com")

	refTwoCookies.SetCookies(d1, []*http.Cookie{c1, c2})

	refDiffDomainCookies.SetCookies(d1, []*http.Cookie{c1})
	refDiffDomainCookies.SetCookies(d2, []*http.Cookie{c2})

	tests := []struct {
		name    string
		args    args
		want    http.CookieJar
		wantErr bool
	}{
		{"No Cookies", args{map[string][]*http.Cookie{}}, refNoCookies, false},
		{"Two Cookies", args{map[string][]*http.Cookie{"https://liferay.com": []*http.Cookie{c1, c2}}}, refTwoCookies, false},
		{"Different Domain Cookies", args{map[string][]*http.Cookie{
			"https://liferay.com":     []*http.Cookie{c1},
			"https://dev.liferay.com": []*http.Cookie{c2},
		}}, refDiffDomainCookies, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fillCookieJar("http://localhost:7010", tt.args.cookieMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCookieJar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCookieJar() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
	{"No rootdomain", args{"", []string{}}, nil, true},
	{"No scheme in rootdomain", args{"google.com", []string{}}, nil, true},
	{"No scheme in otherdomains", args{"https://google.com", []string{"plus.google.com"}}, nil, true},
	{"One otherdomain", args{"https://google.com", []string{"https://plus.google.com"}}, map[string]bool{
		"https://google.com":      true,
		"https://plus.google.com": true,
	}, false},
	{"Two otherdomains", args{"https://google.com", []string{"https://plus.google.com", "https://gmail.com"}}, map[string]bool{
		"https://google.com":      true,
		"https://plus.google.com": true,
		"https://gmail.com":       true,
	}, false},

*/

func Test_getMaxContentLength(t *testing.T) {
	type args struct {
		maxCL int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"Default", args{0}, DefaultMaxContentLength},
		{"Unlimited", args{-1}, UnlimitedMaxContentlength},
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
[cookies]
[[cookies.example]]
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
		Cookies: map[string][]*http.Cookie{
			"example": []*http.Cookie{
				&http.Cookie{
					Domain:  "http://example.com",
					Name:    "Cookie Name",
					Value:   "Cookie Value",
					Path:    "/",
					Expires: date,
				},
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

			if got.RootDomain != tt.want.RootDomain {
				t.Errorf("NewCrawlerFromToml() RootDomain mismatch: %s vs %s", got.RootDomain, tt.want.RootDomain)
				return
			}

			if got.opts.EntryPoint != tt.want.opts.EntryPoint {
				t.Errorf("NewCrawlerFromToml() EntryPoint mismatch: %s vs %s", got.opts.EntryPoint, tt.want.opts.EntryPoint)
				return
			}

			if got.opts.AuthType != tt.want.opts.AuthType {
				t.Errorf("NewCrawlerFromToml() AuthType mismatch: %d vs %d", got.opts.AuthType, tt.want.opts.AuthType)
				return
			}

			if got.opts.User != tt.want.opts.User {
				t.Errorf("NewCrawlerFromToml() User mismatch: %s vs %s", got.opts.User, tt.want.opts.User)
				return
			}

			if got.opts.Pass != tt.want.opts.Pass {
				t.Errorf("NewCrawlerFromToml() Pass mismatch: %s vs %s", got.opts.Pass, tt.want.opts.Pass)
				return
			}

			if got.opts.URLBufferSize != tt.want.opts.URLBufferSize {
				t.Errorf("NewCrawlerFromToml() URLBufferSize mismatch: %d vs %d", got.opts.URLBufferSize, tt.want.opts.URLBufferSize)
				return
			}

			if got.opts.WorkerCount != tt.want.opts.WorkerCount {
				t.Errorf("NewCrawlerFromToml() WorkerCount mismatch: %d vs %d", got.opts.WorkerCount, tt.want.opts.WorkerCount)
				return
			}

			if got.opts.MaxContentLength != tt.want.opts.MaxContentLength {
				t.Errorf("NewCrawlerFromToml() MaxContentLength mismatch: %d vs %d", got.opts.MaxContentLength, tt.want.opts.MaxContentLength)
				return
			}

			if !reflect.DeepEqual(got.opts.AllowedDomains, tt.want.opts.AllowedDomains) {
				t.Errorf("NewCrawlerFromToml() AllowedDomains mismatch: %s vs %s", got.opts.AllowedDomains, tt.want.opts.AllowedDomains)
				return
			}

			if !reflect.DeepEqual(got.opts.Headers, tt.want.opts.Headers) {
				t.Errorf("NewCrawlerFromToml() Headers mismatch: %s vs %s", got.opts.Headers, tt.want.opts.Headers)
				return
			}

			if !reflect.DeepEqual(got.opts.Cookies, tt.want.opts.Cookies) {
				t.Errorf("NewCrawlerFromToml() Cookies mismatch: %s vs %s", got.opts.Cookies, tt.want.opts.Cookies)
				return
			}

			if !reflect.DeepEqual(got.opts.IgnoreGETParameters, tt.want.opts.IgnoreGETParameters) {
				t.Errorf("NewCrawlerFromToml() IgnoreGETParameters mismatch: %s vs %s", got.opts.IgnoreGETParameters, tt.want.opts.IgnoreGETParameters)
				return
			}

			if got.opts.FuzzyGETParameterChecks != tt.want.opts.FuzzyGETParameterChecks {
				t.Errorf("NewCrawlerFromToml() FuzzyGETParameterChecks mismatch: %t vs %t", got.opts.FuzzyGETParameterChecks, tt.want.opts.FuzzyGETParameterChecks)
				return
			}
		})
	}
}
