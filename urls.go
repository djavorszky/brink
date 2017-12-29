package brink

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

func schemeAndHost(_url string) (string, error) {
	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// Link represents a very basic HTML anchor tag
type Link struct {
	Href   string
	Target string
}

// LinksIn expects a valid HTML to parse and returns a slice
// of the links (anchors) contained inside. If "ignoreAnchors"
// is set to true, then links which point to "#someAnchor" type
// locations are ignored.
func LinksIn(response []byte, ignoreAnchors bool) []Link {
	links := make([]Link, 0)

	z := html.NewTokenizer(bytes.NewBuffer(response))
	for {
		if z.Next() == html.ErrorToken {
			// Returning io.EOF indicates success.
			return links
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "a" {
			var l Link
			for _, attr := range t.Attr {
				switch attr.Key {
				case "href":
					l.Href = attr.Val
				case "target":
					l.Target = attr.Val
				}
			}

			if l.Href == "javascript:;" ||
				(ignoreAnchors && strings.HasPrefix(l.Href, "#")) {
				continue
			}

			links = append(links, l)
		}
	}
}

// SortGetParameters expects a full URL and returns one in which the GET parameters
// have been sorted by their keys.
func SortGetParameters(_url string) (string, error) {
	var result []string

	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}
	params := u.Query()

	if len(params) == 0 {
		return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
	}

	for key, vals := range params {
		for _, val := range vals {
			if val == "" {
				result = append(result, key)
				continue
			}

			result = append(result, fmt.Sprintf("%s=%s", key, val))
		}
	}

	sort.Strings(result)

	return fmt.Sprintf("%s://%s?%s", u.Scheme, u.Host, strings.Join(result, "&")), nil
}
