package main

import (
	"cut"
	"flag"
	"fmt"
	"os"
)

func main() {
	fieldsFlag := flag.String("f", "", "number of fields")
	delimiterFlag := flag.String("d", "\t", "delimiter")
	separatedFlag := flag.Bool("s", false, "separated")

	flag.Parse()

	if *fieldsFlag == "" {
		fmt.Fprintln(os.Stderr, "flag -f is required")
		os.Exit(1)
	}

	fields, err := cut.ParseFields(*fieldsFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse fields:", err)
		os.Exit(1)
	}

	opts := cut.Options{
		Fields:    fields,
		Delimiter: *delimiterFlag,
		Separated: *separatedFlag,
	}

	if err := cut.Cut(os.Stdin, os.Stdout, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
