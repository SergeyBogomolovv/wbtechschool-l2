package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func main() {
	var check, unique bool
	var opts Options

	flag.BoolVar(&unique, "u", false, "output only unique values.")
	flag.BoolVar(&check, "c", false, "check whether input is sorted; do not sort.")
	flag.IntVar(&opts.Column, "k", 0, "sort by field N (1-based). Fields are TAB-separated by default.")
	flag.BoolVar(&opts.Numeric, "n", false, "compare according to numeric value.")
	flag.BoolVar(&opts.Reverse, "r", false, "reverse the result of comparisons.")
	flag.BoolVar(&opts.Month, "M", false, "compare by month name (Jan…Dec).")
	flag.BoolVar(&opts.IgnoreTrailingBlanks, "b", false, "ignore trailing blanks when comparing.")
	flag.BoolVar(&opts.HumanNumeric, "h", false, "compare human-readable numbers (e.g., 2K, 3M).")

	os.Args = unwrapFlags(os.Args)
	flag.Parse()

	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "No such file or directory")
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	if check {
		ok, unsorted := checkSorted(in, cmpFunc(opts))
		if !ok {
			fmt.Fprintf(os.Stderr, "unordered: %s\n", unsorted)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if unique {
		uw := newUniqueWriter(out)
		defer uw.Flush()
		out = uw
	}

	if err := ExternalSort(in, out, cmpFunc(opts)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// функция для обработки флагов -ur => -u -r
func unwrapFlags(args []string) []string {
	var expanded []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") &&
			!strings.HasPrefix(arg, "--") &&
			len(arg) > 2 {

			if strings.Contains(arg, "=") {
				expanded = append(expanded, arg)
				continue
			}

			if arg[1] != '-' && len(arg) >= 3 && unicode.IsDigit(rune(arg[2])) {
				expanded = append(expanded, arg)
				continue
			}

			for _, ch := range arg[1:] {
				expanded = append(expanded, "-"+string(ch))
			}
		} else {
			expanded = append(expanded, arg)
		}
	}
	return expanded
}
