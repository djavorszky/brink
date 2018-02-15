package brink

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

func schemeAndHost(_url string) (string, string, error) {
	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", "", fmt.Errorf("failed parsing url: %v", err)
	}

	return u.Scheme, u.Host, nil
}

func scheme(_url string) (string, error) {
	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}

	return u.Scheme, nil
}

// Link represents a very basic HTML anchor tag. LinkedFrom is the page on which it is found,
// Href is where it is pointing to.
type Link struct {
	LinkedFrom string
	Href       string
	Target     string
}

// AbsoluteLinksIn expects a valid HTML to parse and returns a slice
// of the links (anchors) contained inside. If "ignoreAnchors"
// is set to true, then links which point to "#someAnchor" type
// locations are ignored.
//
// If any links within the HTML start with a forward slash (e.g. is a dynamic link),
// it will get prepended with the passed url.
func AbsoluteLinksIn(hostURL, linkedFrom string, body []byte, ignoreAnchors bool) ([]Link, error) {
	scheme, host, err := schemeAndHost(hostURL)
	if err != nil {
		return nil, fmt.Errorf("failed parsing url: %v", err)
	}

	links := LinksIn(linkedFrom, body, ignoreAnchors)
	for ix, l := range links {
		if strings.HasPrefix(l.Href, "//") {
			l.Href = fmt.Sprintf("%s://%s", scheme, l.Href)
			links[ix] = l
		}

		if strings.HasPrefix(l.Href, "/") {
			l.Href = fmt.Sprintf("%s://%s%s", scheme, host, l.Href)
			links[ix] = l
		}
	}

	return links, nil
}

// LinksIn expects a valid HTML to parse and returns a slice
// of the links (anchors) contained inside. If "ignoreAnchors"
// is set to true, then links which point to "#someAnchor" type
// locations are ignored.
func LinksIn(linkedFrom string, body []byte, ignoreAnchors bool) []Link {
	links := make([]Link, 0)

	z := html.NewTokenizer(bytes.NewBuffer(body))
	for {
		if z.Next() == html.ErrorToken {
			// Returning io.EOF indicates success.
			return links
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "a" {
			l := Link{LinkedFrom: linkedFrom}
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

// normalizeURL expects a full URL and returns one in which the GET parameters
// have been sorted by their keys. It also removes any GET parameters which the
// Crawler has been told to ignore.
func (c *Crawler) normalizeURL(_url string) (string, error) {
	var (
		result       []string
		ignoreParams bool
	)

	u, err := url.ParseRequestURI(strings.TrimSpace(_url))
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}
	params := u.Query()

	if c.ignoredGETParams.Size() > 0 {
		ignoreParams = true
	}

	for key, vals := range params {
		if ignoreParams {
			if c.ignoredGETParams.Contains(key) {
				continue
			}

			if c.opts.FuzzyGETParameterChecks && c.ignoredGETParams.AnyContainsReverse(key) {
				continue
			}
		}

		for _, val := range vals {
			if val == "" {
				result = append(result, key)
				continue
			}

			result = append(result, fmt.Sprintf("%s=%s", key, val))
		}
	}

	if len(result) == 0 {
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path), nil
	}

	sort.Strings(result)

	return fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, strings.Join(result, "&")), nil
}
