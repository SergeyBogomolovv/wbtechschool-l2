package main

import (
	"bufio"
	"io"
)

// проверка, что входные данные отсортированы
func checkSorted(in io.Reader, cmp func(a, b string) int) (bool, string) {
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	prev := scanner.Text()
	for scanner.Scan() {
		if cmp(prev, scanner.Text()) > 0 {
			return false, prev
		}
		prev = scanner.Text()
	}

	return true, ""
}
