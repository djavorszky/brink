package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/djavorszky/brink"
	"golang.org/x/net/html"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: brink url")
		os.Exit(1)
	}

	cookies := make(map[string][]*http.Cookie)

	cookies[os.Args[1]] = []*http.Cookie{
		&http.Cookie{Domain: ".liferay.int", Name: "user", Value: "daniel.javorszky@liferay.com"},
	}

	c, err := brink.NewCrawlerWithOpts(os.Args[1], brink.CrawlOptions{Cookies: cookies})
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		os.Exit(1)
	}

	status, body, err := c.Crawl(os.Args[1])
	if err != nil {
		fmt.Printf("oops: %v", err)
		os.Exit(1)
	}

	fmt.Printf("status code: %d", status)

	z := html.NewTokenizer(bytes.NewBuffer(body))
	for {
		if z.Next() == html.ErrorToken {
			// Returning io.EOF indicates success.
			return
		}

		t := z.Token()
		fmt.Printf("Type: %s, Data: %s\n", t.Type, t.Data)
	}
}
