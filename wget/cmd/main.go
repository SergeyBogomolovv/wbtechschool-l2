package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: wget <url>")
		os.Exit(1)
	}
	url := os.Args[1]

	_ = url
}
