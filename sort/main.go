package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type Options struct {
	Sort SortOptions

	Unique bool
	Check  bool

	FilePath string
}

func main() {
	opts := parseOpts(expandArgs(os.Args[1:]))

	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	if opts.FilePath != "" {
		f, err := os.Open(opts.FilePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	if opts.Check {
		ok, unsorted := checkSorted(in, cmpFunc(opts.Sort))
		if !ok {
			fmt.Fprintf(os.Stderr, "sort: disorder: %s\n", unsorted)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if opts.Unique {
		uw := newUniqueWriter(out)
		defer uw.Flush()
		out = uw
	}

	if err := ExternalSort(in, out, cmpFunc(opts.Sort)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOpts(args []string) Options {
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)

	var opts Options
	fs.BoolVar(&opts.Unique, "u", false, "output only unique values.")
	fs.BoolVar(&opts.Check, "c", false, "check whether input is sorted; do not sort.")
	fs.IntVar(&opts.Sort.Column, "k", 0, "sort by field N (1-based). Fields are TAB-separated by default.")
	fs.BoolVar(&opts.Sort.Numeric, "n", false, "compare according to numeric value.")
	fs.BoolVar(&opts.Sort.Reverse, "r", false, "reverse the result of comparisons.")
	fs.BoolVar(&opts.Sort.Month, "M", false, "compare by month name (Jan…Dec).")
	fs.BoolVar(&opts.Sort.IgnoreTrailingBlanks, "b", false, "ignore trailing blanks when comparing.")
	fs.BoolVar(&opts.Sort.HumanNumeric, "h", false, "compare human-readable numbers (e.g., 2K, 3M).")

	fs.Parse(args)

	if fs.NArg() > 0 {
		opts.FilePath = fs.Arg(0)
	}

	return opts
}

func expandArgs(args []string) []string {
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
