package brink

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"testing"

	"github.com/djavorszky/brink/store"
	"golang.org/x/net/publicsuffix"
)

func TestNewCrawler(t *testing.T) {
	testCrawler := &Crawler{
		RootDomain:     "https://liferay.com",
		allowedDomains: store.New(),
		visitedURLs:    store.New(),
		handlers:       make(map[int]func(url string, status int, body string)),
		client:         &http.Client{},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCrawler() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			got, err := fillCookieJar(tt.args.cookieMap)
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
