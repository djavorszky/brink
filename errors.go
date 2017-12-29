package brink

import "fmt"

// NotAllowed error is returned by Fetch when a domain is not allowed
// to be visited.
type NotAllowed struct {
	domain string
}

func (na NotAllowed) Error() string {
	return fmt.Sprintf("domain not allowed: %v", na.domain)
}

// ContentTooLarge error is returned by Fetch when the content length
// of a page is larger than what is allowed.
type ContentTooLarge struct {
	url string
}

func (ctl ContentTooLarge) Error() string {
	return fmt.Sprintf("content-length too large of url: %v", ctl.url)
}
