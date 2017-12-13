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
