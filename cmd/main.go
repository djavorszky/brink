package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/djavorszky/brink"
)

func main() {
	config := flag.String("conf", "brink.toml", "Specify the configuration filename to be used")

	flag.Parse()

	c, err := brink.NewCrawlerFromToml(*config)
	if err != nil {
		fmt.Printf("Failed initializing crawler: %v\n", err)
		os.Exit(1)
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		c.Stop()
	}()

	c.HandleDefaultFunc(handler)
	c.HandleFunc(http.StatusNotFound, notFoundHandler)

	c.Start()
}

var oks int

func handler(linkedFrom, url string, status int, body string, cached bool) {
	oks++

	if oks%100 == 0 {
		log.Printf("Links seen: %d", oks)
	}

	if cached {
		return
	}

	if strings.Contains(body, "Use the buttons below to create it or to search for the words in the title.") {
		log.Printf("%s -> %s: linked wiki article does not exist", linkedFrom, url)
	}
}

func notFoundHandler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		log.Printf("404: CACHED: %s -> %s", linkedFrom, url)
	} else {
		log.Printf("404: %s -> %s", linkedFrom, url)
	}
}
