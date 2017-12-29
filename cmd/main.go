package main

import (
	"fmt"
	"os"

	"github.com/djavorszky/brink"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: brink url")
		os.Exit(1)
	}

	c, err := brink.NewCrawlerWithOpts(os.Args[1], brink.CrawlOptions{})
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		os.Exit(1)
	}

	_, body, err := c.Fetch(os.Args[1])
	if err != nil {
		fmt.Printf("oops: %v", err)
		os.Exit(1)
	}

	links := brink.LinksIn(body, true)

	for _, link := range links {
		fmt.Printf("To: %v\n", link.Href)
	}

	fmt.Println(len(links))

}
