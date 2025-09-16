package storage

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Storage - хранилище
type Storage struct {
	Root string
}

// New - Конструктор
func New(root string) *Storage {
	return &Storage{Root: root}
}

// LocalPath - возвращает путь на диске для данного URL
func (s *Storage) LocalPath(u *url.URL) string {
	host := u.Hostname()

	path := u.Path
	if path == "" || path == "/" {
		path = "/index.html"
	}

	if strings.HasSuffix(path, "/") {
		path = filepath.Join(path, "index.html")
	}

	path = strings.TrimPrefix(path, "/")

	ext := filepath.Ext(path)
	if ext == "" {
		path = path + ".html"
	}

	if u.RawQuery != "" {
		ext = filepath.Ext(path)
		base := strings.TrimSuffix(path, ext)
		path = base + "_q=" + sanitize(u.RawQuery) + ext
	}

	local := filepath.Join(s.Root, host, path)
	return local
}

// sanitize - приводит строку к безопасному для имени файла виду
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// Save - сохраняет файл по пути
func (s *Storage) Save(u *url.URL, data []byte) (string, error) {
	localPath := s.LocalPath(u)

	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", fmt.Errorf("write file %s: %w", localPath, err)
	}

	return localPath, nil
}
