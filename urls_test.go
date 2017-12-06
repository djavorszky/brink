package brink

import "testing"

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
