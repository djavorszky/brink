package brink

import (
	"bytes"
	"fmt"
	"net/url"

	"golang.org/x/net/html"
)

type link struct {
	Href   string
	Target string
}

func schemeAndHost(_url string) (string, error) {
	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

func LinksIn(response []byte) []link {
	links := make([]link, 0)

	z := html.NewTokenizer(bytes.NewBuffer(response))
	for {
		if z.Next() == html.ErrorToken {
			// Returning io.EOF indicates success.
			return links
		}

		t := z.Token()

		if t.Type == html.StartTagToken && t.Data == "a" {
			var l link
			for _, attr := range t.Attr {
				switch attr.Key {
				case "href":
					l.Href = attr.Val
				case "target":
					l.Target = attr.Val
				}
			}

			links = append(links, l)
		}
	}
}
