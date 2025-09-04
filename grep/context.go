package grep

import (
	"fmt"
	"io"
)

type line struct {
	num  int
	text string
}

// ContextPrinter - отвечает за вывод контекста без повторений
type ContextPrinter struct {
	w           io.Writer
	showLine    bool
	before      int
	after       int
	ring        []line
	pos         int
	filled      int
	postCount   int
	lastPrinted int
}

// NewContextPrinter - создает новый ContextPrinter
func NewContextPrinter(w io.Writer, before, after int, showLine bool) *ContextPrinter {
	cp := &ContextPrinter{w: w, showLine: showLine, before: before, after: after}
	if before > 0 {
		cp.ring = make([]line, before)
	}
	return cp
}

// Handle - выводит строку и контекст, без повторений
func (cp *ContextPrinter) Handle(num int, s string, isMatch bool) {
	if isMatch {
		if cp.postCount == 0 && cp.before > 0 && cp.filled > 0 {
			for i := cp.filled; i > 0; i-- {
				idx := (cp.pos - i + cp.before) % cp.before
				cp.printLine(cp.ring[idx].num, cp.ring[idx].text)
			}
		}
		cp.printLine(num, s)
		cp.postCount = cp.after
	} else if cp.postCount > 0 {
		cp.printLine(num, s)
		cp.postCount--
	}
	cp.push(num, s)
}

func (cp *ContextPrinter) printLine(n int, s string) {
	if n == cp.lastPrinted {
		return
	}
	cp.lastPrinted = n
	if cp.showLine {
		fmt.Fprintf(cp.w, "%d:%s\n", n, s)
	} else {
		fmt.Fprintln(cp.w, s)
	}
}

func (cp *ContextPrinter) push(n int, s string) {
	if cp.before == 0 {
		return
	}
	cp.ring[cp.pos] = line{num: n, text: s}
	cp.pos = (cp.pos + 1) % cp.before
	if cp.filled < cp.before {
		cp.filled++
	}
}
