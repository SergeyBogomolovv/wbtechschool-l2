package sort

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// Options - опции для команды sort
type Options struct {
	Column               int
	Numeric              bool
	Month                bool
	Reverse              bool
	IgnoreTrailingBlanks bool
	HumanNumeric         bool

	Unique bool
	Check  bool
}

// Sort - Сортировка больших обьемов строк.
func Sort(in io.Reader, out io.Writer, opts Options) error {
	cmp := newCmpFunc(opts)
	if opts.Check {
		ok, unsorted := checkSorted(in, opts.IgnoreTrailingBlanks, cmp)
		if !ok {
			return fmt.Errorf("sort: disorder: %s", unsorted)
		}
		return nil
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "extsort-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	files, err := splitChunks(in, tmpDir, opts.IgnoreTrailingBlanks, cmp)
	if err != nil {
		return err
	}

	if opts.Unique {
		writer := newUniqueWriter(out)
		defer writer.Flush()
		out = writer
	}

	return sortFiles(files, out, cmp)
}

// разбивает входные данные на чанки, сортирует их и сохраняет во временные файлы
func splitChunks(in io.Reader, tmpDir string, trimTrailingBlanks bool, cmp func(a, b string) int) ([]string, error) {
	const maxLines = 100000

	lines := make([]string, 0)
	files := make([]string, 0)

	flush := func() error {
		if len(lines) == 0 {
			return nil
		}
		slices.SortFunc(lines, cmp)

		file, err := os.CreateTemp(tmpDir, "chunk-*.txt")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		for _, s := range lines {
			fmt.Fprintln(writer, s)
		}

		files = append(files, file.Name())
		lines = lines[:0]
		return nil
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if trimTrailingBlanks {
			line = strings.TrimRight(line, " \t\r")
		}
		lines = append(lines, line)
		if len(lines) >= maxLines {
			if err := flush(); err != nil {
				return files, err
			}
		}
	}

	if err := flush(); err != nil {
		return files, err
	}

	return files, scanner.Err()
}

// сортирует данные из отсротиванных чанков
func sortFiles(files []string, out io.Writer, cmp func(a, b string) int) error {
	if len(files) == 0 {
		return nil
	}

	// открытие всех чанков
	scanners := make([]*bufio.Scanner, 0, len(files))
	for _, p := range files {
		file, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		scanners = append(scanners, bufio.NewScanner(file))
	}

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// инициализация кучи
	h := &mergeHeap{cmp: cmp}
	for id, scanner := range scanners {
		if !scanner.Scan() {
			continue
		}
		line := scanner.Text()
		h.items = append(h.items, mergeItem{line: line, fileID: id})
	}
	heap.Init(h)

	// сортировка
	for h.Len() > 0 {
		el := heap.Pop(h).(mergeItem)
		fmt.Fprintln(out, el.line)
		if !scanners[el.fileID].Scan() {
			continue
		}
		next := scanners[el.fileID].Text()
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

// проверка, что входные данные отсортированы
func checkSorted(in io.Reader, ignoreTrailingBlanks bool, cmp func(a, b string) int) (bool, string) {
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	prev := scanner.Text()
	for scanner.Scan() {
		line := scanner.Text()
		if ignoreTrailingBlanks {
			line = strings.TrimRight(line, " \t\r")
		}
		if cmp(prev, line) > 0 {
			return false, prev
		}
		prev = scanner.Text()
	}

	return true, ""
}
