package brink

import (
	"fmt"
	"net/url"
)

func schemeAndHost(_url string) (string, error) {
	u, err := url.ParseRequestURI(_url)
	if err != nil {
		return "", fmt.Errorf("failed parsing url: %v", err)
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}
