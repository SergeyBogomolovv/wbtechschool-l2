package main

import (
	"flag"
	"fmt"
	"io"
	sort "mysort"
	"os"
	"strings"
	"unicode"
)

func main() {
	args := expandArgs(os.Args[1:])
	filePath, opts, err := parseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var in io.Reader = os.Stdin

	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	if err := sort.Sort(in, os.Stdout, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOptions(args []string) (string, sort.Options, error) {
	var opts sort.Options
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)

	fs.BoolVar(&opts.Unique, "u", false, "output only unique values.")
	fs.BoolVar(&opts.Check, "c", false, "check whether input is sorted; do not sort.")
	fs.IntVar(&opts.Column, "k", 0, "sort by field N (1-based). Fields are TAB-separated by default.")
	fs.BoolVar(&opts.Numeric, "n", false, "compare according to numeric value.")
	fs.BoolVar(&opts.Reverse, "r", false, "reverse the result of comparisons.")
	fs.BoolVar(&opts.Month, "M", false, "compare by month name (Jan…Dec).")
	fs.BoolVar(&opts.IgnoreTrailingBlanks, "b", false, "ignore trailing blanks when comparing.")
	fs.BoolVar(&opts.HumanNumeric, "h", false, "compare human-readable numbers (e.g., 2K, 3M).")

	if err := fs.Parse(args); err != nil {
		return "", opts, err
	}

	var filePath string
	if fs.NArg() > 0 {
		filePath = fs.Arg(0)
	}

	return filePath, opts, nil
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
