package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/djavorszky/brink"
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

	_, body, err := c.Fetch(os.Args[1])
	if err != nil {
		fmt.Printf("oops: %v", err)
		os.Exit(1)
	}

	links := brink.LinksIn(body)

	for _, link := range links {
		fmt.Printf("To: %v\n", link.Href)
	}

	fmt.Println(len(links))
}
