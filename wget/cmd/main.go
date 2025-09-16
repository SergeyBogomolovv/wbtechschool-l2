package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"wget/internal/crawler"
	"wget/internal/downloader"
	"wget/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: crawler <url> [depth]")
		os.Exit(1)
	}

	rawURL := os.Args[1]
	depth := 1
	if len(os.Args) >= 3 {
		var err error
		depth, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Invalid depth:", os.Args[2])
			os.Exit(1)
		}
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Invalid URL:", rawURL)
		os.Exit(1)
	}

	// Инициализация компонентов
	downloader := downloader.New(10 * time.Second)
	storage := storage.New("wget-output")

	crawler := crawler.New(u, depth, downloader, storage)

	if err := crawler.Run(); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Done. Saved to ./wget-output")
	}
}
