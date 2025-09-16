package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Downloader - скачивает данные страниц
type Downloader struct {
	client *http.Client
}

// New - Конструктор с таймаутом
func New(timeout time.Duration) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Fetch скачивает данные по URL
func (d *Downloader) Fetch(u *url.URL) ([]byte, string, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	// Задаём User-Agent, чтобы сайты не блокировали
	req.Header.Set("User-Agent", "GoCrawler/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("http get %s: %w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status %d for %s", resp.StatusCode, u)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	return body, contentType, nil
}
