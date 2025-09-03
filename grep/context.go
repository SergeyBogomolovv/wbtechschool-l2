package main

import (
	"fmt"
	"io"
)

type priorLine struct {
	num  int
	text string
}

type contextPrinter struct {
	w           io.Writer
	showLine    bool
	before      int
	after       int
	ring        []priorLine
	rpos        int
	rfilled     int
	postCount   int
	lastPrinted int
}

func newContextPrinter(w io.Writer, before, after int, showLine bool) *contextPrinter {
	cp := &contextPrinter{w: w, showLine: showLine, before: before, after: after}
	if before > 0 {
		cp.ring = make([]priorLine, before)
	}
	return cp
}

func (cp *contextPrinter) handle(lineNum int, line string, isMatch bool) {
	if isMatch {
		if cp.postCount == 0 && cp.before > 0 && cp.rfilled > 0 {
			for i := cp.rfilled; i > 0; i-- {
				idx := (cp.rpos - i + cp.before) % cp.before
				cp.printLine(cp.ring[idx].num, cp.ring[idx].text)
			}
		}
		cp.printLine(lineNum, line)
		cp.postCount = cp.after
	} else if cp.postCount > 0 {
		cp.printLine(lineNum, line)
		cp.postCount--
	}
	cp.push(lineNum, line)
}

func (cp *contextPrinter) printLine(n int, s string) {
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

func (cp *contextPrinter) push(n int, s string) {
	if cp.before == 0 {
		return
	}
	cp.ring[cp.rpos] = priorLine{num: n, text: s}
	cp.rpos = (cp.rpos + 1) % cp.before
	if cp.rfilled < cp.before {
		cp.rfilled++
	}
}
