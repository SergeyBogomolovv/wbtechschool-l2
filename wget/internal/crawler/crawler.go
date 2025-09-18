package crawler

import (
	"fmt"
	"mime"
	"net/url"
	"sync"
	"wget/internal/parser"
)

// Downloader - интерфейс для работы с загрузчиком
type Downloader interface {
	Fetch(u *url.URL) ([]byte, string, error) // Скачать ресурс, вернуть тело и Content-Type
}

// Storage - интерфейс для работы с хранилищем
type Storage interface {
	LocalPath(u *url.URL) string                  // Получить путь на диске для URL
	Save(u *url.URL, data []byte) (string, error) // Сохранить файл
}

// Task - одно задание в очереди
type Task struct {
	URL   *url.URL
	Depth int
}

// Crawler - основная структура программы
type Crawler struct {
	startURL   *url.URL
	maxDepth   int
	downloader Downloader
	storage    Storage

	visited   map[string]bool
	mu        sync.Mutex
	queue     chan Task
	tasksWg   sync.WaitGroup
	workersWg sync.WaitGroup
}

// New - Конструктор
func New(start *url.URL, depth int, d Downloader, s Storage) *Crawler {
	return &Crawler{
		startURL:   start,
		maxDepth:   depth,
		downloader: d,
		storage:    s,
		visited:    make(map[string]bool),
		queue:      make(chan Task, 100),
	}
}

// Run - Запускает программу
func (c *Crawler) Run() error {
	c.enqueue(c.startURL, c.maxDepth)

	go func() {
		c.tasksWg.Wait()
		close(c.queue)
	}()

	const numWorkers = 4
	c.workersWg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer c.workersWg.Done()
			for task := range c.queue {
				c.process(task)
			}
		}()
	}

	c.workersWg.Wait()
	return nil
}

func (c *Crawler) enqueue(u *url.URL, depth int) {
	n := normalizeURL(u)
	if n.Hostname() != c.startURL.Hostname() {
		return
	}

	c.mu.Lock()
	if c.visited[n.String()] {
		c.mu.Unlock()
		return
	}
	c.visited[n.String()] = true
	c.mu.Unlock()

	if depth < 0 {
		return
	}

	c.tasksWg.Add(1)
	c.queue <- Task{URL: n, Depth: depth}
}

func (c *Crawler) process(t Task) {
	fmt.Println("Downloading:", t.URL)
	defer c.tasksWg.Done()

	body, contentType, err := c.downloader.Fetch(t.URL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// если HTML — парсим
	if t.Depth > 0 && isHTML(contentType) {
		links, err := parser.ExtractLinks(t.URL, body)
		if err == nil {
			for _, link := range links {
				c.enqueue(link, t.Depth-1)
			}
		}
		// Переписываем ссылки под локальные пути
		body, err = parser.RewriteLinks(t.URL, body, c.storage.LocalPath)
		if err != nil {
			fmt.Println("RewriteLinks error:", err)
		}
	}

	_, err = c.storage.Save(t.URL, body)
	if err != nil {
		fmt.Println("Save error:", err)
	}
}

func isHTML(ct string) bool {
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return mediaType == "text/html"
}

func normalizeURL(u *url.URL) *url.URL {
	clone := *u
	clone.Fragment = ""
	if clone.Path == "" {
		clone.Path = "/"
	}
	return &clone
}
