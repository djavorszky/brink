package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		c.Stop()
	}()

	c.HandleDefaultFunc(handler)

	c.Start()
}

func handler(url string, status int, body string) {
	log.Printf("status: %d, url: %v", status, url)
}
