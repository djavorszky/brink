package brink

import (
	"reflect"
	"testing"
)

func Test_schemeAndHost(t *testing.T) {
	type args struct {
		_url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Valid URL", args{"https://google.com"}, "https://google.com", false},
		{"Valid URL w/ port", args{"https://google.com:80"}, "https://google.com:80", false},
		{"Valid URL w/ trailing slash", args{"https://google.com/"}, "https://google.com", false},
		{"Valid URL w/ paths", args{"https://google.com/some/path"}, "https://google.com", false},
		{"Valid URL w/ paths && trailing slash", args{"https://google.com/some/path/"}, "https://google.com", false},

		{"Invalid scheme", args{"https//google.com"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := schemeAndHost(tt.args._url)
			if (err != nil) != tt.wantErr {
				t.Errorf("schemeAndHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("schemeAndHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_LinksIn(t *testing.T) {
	type args struct {
		response     []byte
		ignoreAnchor bool
	}
	tests := []struct {
		name string
		args args
		want []Link
	}{
		{"no links no anchors", args{[]byte("<html><header><title>This is title</title></header><body>Hello world</body></html>"), false}, []Link{}},
		{"no links with anchors", args{[]byte("<html><header><title>This is title</title></header><body>Hello world</body></html>"), true}, []Link{}},
		{"one link with anchors",
			args{[]byte("<html><header><title>This is title</title></header><body><a href=\"#\">Hello world</a></body></html>"), false},
			[]Link{Link{Href: "#"}},
		},
		{"ignore anchor",
			args{[]byte("<html><header><title>This is title</title></header><body><a href=\"#\">Hello world</a></body></html>"), true},
			[]Link{},
		},
		{"one link with target blank",
			args{[]byte("<html><header><title>This is title</title></header><body><a href=\"google.com\" target=\"_blank\">Hello world</a></body></html>"), true},
			[]Link{Link{Href: "google.com", Target: "_blank"}},
		},
		{"two links with target blank",
			args{[]byte("<html><header><title>This is title</title></header><body><a href=\"google.com\">Hello world</a><a href=\"liferay.com\" target=\"_blank\">Whatsup</a></body></html>"), true},
			[]Link{
				Link{Href: "google.com"},
				Link{Href: "liferay.com", Target: "_blank"},
			},
		},
		{"one link with javascript",
			args{[]byte("<html><header><title>This is title</title></header><body><a href=\"javascript:;\">Hello world</a></body></html>"), false},
			[]Link{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinksIn(tt.args.response, tt.args.ignoreAnchor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizeURL(t *testing.T) {
	normCrawler, _ := NewCrawler("https://liferay.com")
	ignoreCrawler, _ := NewCrawlerWithOpts("https://liferay.com", CrawlOptions{IgnoreGETParameters: []string{"something"}})

	type args struct {
		c    *Crawler
		_url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no_get_params", args{normCrawler, "https://liferay.com"}, "https://liferay.com"},
		{"one_get_param", args{normCrawler, "https://liferay.com?test=something"}, "https://liferay.com?test=something"},
		{"two_get_params", args{normCrawler, "https://liferay.com?test=justTesting&something=123"}, "https://liferay.com?something=123&test=justTesting"},
		{"one_get_param_no_value", args{normCrawler, "https://liferay.com?test"}, "https://liferay.com?test"},
		{"two_get_params_no_value", args{normCrawler, "https://liferay.com?test&something"}, "https://liferay.com?something&test"},
		{"one_get_param_not_ignored", args{ignoreCrawler, "https://liferay.com?test=something"}, "https://liferay.com?test=something"},
		{"one_get_param_ignored", args{ignoreCrawler, "https://liferay.com?something=test"}, "https://liferay.com"},
		{"two_get_params_none_ignored", args{ignoreCrawler, "https://liferay.com?test=justTesting&shoot=123"}, "https://liferay.com?shoot=123&test=justTesting"},
		{"two_get_params_one_ignored", args{ignoreCrawler, "https://liferay.com?test=justTesting&something=123"}, "https://liferay.com?test=justTesting"},
		{"two_get_params_both_ignored", args{ignoreCrawler, "https://liferay.com?something=justTesting&something=123"}, "https://liferay.com"},
		{"one_get_param_no_value_not_ignored", args{ignoreCrawler, "https://liferay.com?test"}, "https://liferay.com?test"},
		{"one_get_param_no_value_ignored", args{ignoreCrawler, "https://liferay.com?something"}, "https://liferay.com"},
		{"two_get_params_no_value_one_ignored", args{ignoreCrawler, "https://liferay.com?test&something"}, "https://liferay.com?test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.c.normalizeURL(tt.args._url); got != tt.want {
				t.Errorf("SortGetParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
