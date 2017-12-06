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

	w, err := brink.NewWalker(os.Args[1])
	if err != nil {
		fmt.Printf("oops: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(w)
}
