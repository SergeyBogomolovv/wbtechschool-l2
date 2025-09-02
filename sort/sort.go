package main

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// ExternalSort - Сортировка больших обьемов строк
func ExternalSort(in io.Reader, out io.Writer, cmp func(a, b string) int) error {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "extsort-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	files, err := splitChunks(in, tmpDir, cmp)
	if err != nil {
		return err
	}

	return merge(files, out, cmp)
}

// разбивает входные данные на чанки, сортирует их и сохраняет во временные файлы
func splitChunks(in io.Reader, dir string, cmp func(a, b string) int) ([]string, error) {
	const (
		chunkSize  = 256 << 20 // 256MB
		bufferSize = 1 << 20   // 1MB
	)

	lines := make([]string, 0)
	files := make([]string, 0)
	accumBytes := 0

	flush := func() error {
		if len(lines) == 0 {
			return nil
		}
		slices.SortFunc(lines, cmp)

		file, err := os.CreateTemp(dir, "chunk-*.txt")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer file.Close()

		writer := bufio.NewWriterSize(file, bufferSize)
		defer writer.Flush()

		for _, s := range lines {
			_, err := writer.WriteString(s)
			if err != nil {
				return fmt.Errorf("failed to write to temp file: %w", err)
			}
			err = writer.WriteByte('\n')
			if err != nil {
				return fmt.Errorf("failed to write to temp file: %w", err)
			}
		}

		files = append(files, file.Name())
		lines = lines[:0]
		accumBytes = 0
		return nil
	}

	reader := bufio.NewReaderSize(in, bufferSize)
	for {
		b, err := reader.ReadBytes('\n')
		if len(b) > 0 {
			if accumBytes+len(b) > chunkSize && len(lines) > 0 {
				if err := flush(); err != nil {
					return files, err
				}
			}
			accumBytes += len(b)
			s := strings.TrimRight(string(b), "\r\n")
			lines = append(lines, s)
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return files, err
		}
	}

	if err := flush(); err != nil {
		return files, err
	}

	return files, nil
}

// сортирует данные из отсротиванных чанков
func merge(paths []string, out io.Writer, cmp func(a, b string) int) error {
	const bufferSize = 1 << 20 // 1MB

	if len(paths) == 0 {
		return nil
	}

	// открытие всех чанков
	readers := make([]*bufio.Reader, 0, len(paths))
	for _, p := range paths {
		file, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		readers = append(readers, bufio.NewReaderSize(file, bufferSize))
	}

	writer := bufio.NewWriterSize(out, bufferSize)
	defer writer.Flush()

	// инициализация кучи
	h := &mergeHeap{cmp: cmp}
	for i, reader := range readers {
		line, err := readLine(reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to read line: %w", err)
		}
		h.items = append(h.items, mergeItem{line: line, fileID: i})
	}
	heap.Init(h)

	// сортировка
	for h.Len() > 0 {
		el := heap.Pop(h).(mergeItem)
		if _, err := writer.WriteString(el.line); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
		if err := writer.WriteByte('\n'); err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
		next, err := readLine(readers[el.fileID])
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to read line: %w", err)
		}
		heap.Push(h, mergeItem{line: next, fileID: el.fileID})
	}

	return nil
}

type mergeItem struct {
	line   string
	fileID int
}

type mergeHeap struct {
	items []mergeItem
	cmp   func(a, b string) int
}

func (h mergeHeap) Len() int { return len(h.items) }
func (h mergeHeap) Less(i, j int) bool {
	return h.cmp(h.items[i].line, h.items[j].line) < 0
}
func (h mergeHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *mergeHeap) Push(x any)   { h.items = append(h.items, x.(mergeItem)) }
func (h *mergeHeap) Pop() any {
	n := len(h.items)
	it := h.items[n-1]
	h.items = h.items[:n-1]
	return it
}

func readLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if errors.Is(err, io.EOF) && len(s) > 0 {
		return strings.TrimRight(s, "\r\n"), nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}
