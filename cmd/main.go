package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	//c.HandleFunc(http.StatusNotFound, notFoundHandler)

	c.Start()
}

func handler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		log.Printf("%s -> %s: %d, cached", linkedFrom, url, status)
	} else {
		log.Printf("%s -> %s: %d", linkedFrom, url, status)
	}

	log.Println(body)
}

/*
var (
	count    int
	statuses = make(map[int]int, 0)
)
func handler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		return
	}

	count++
	statuses[status] = statuses[status] + 1

	if count%100 == 0 {
		statusStr := ""
		for status, statusCount := range statuses {
			statusStr = fmt.Sprintf("%s status %d: %d", statusStr, status, statusCount)
		}

		log.Println(statusStr)
	}
}

func notFoundHandler(linkedFrom, url string, status int, body string, cached bool) {
	if cached {
		log.Printf("CACHED: %s -> %s: 404", linkedFrom, url)
	} else {
		log.Printf("%s -> %s: 404", linkedFrom, url)
	}

	count++
	statuses[status] = statuses[status] + 1
}
*/
