package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Options - опции для команды grep
type Options struct {
	// количество строк после
	After int
	// количество строк до
	Before int
	// количество строк до и после
	Context int

	// вывести только количество найденных строк
	CountLines bool
	// игнорировать регистр
	IgnoreCase bool
	// инвертировать условие
	Invert bool
	// выводить номер строки
	ShowLine bool
	// воспринимать шаблон как фиксированную строку, а не регулярное выражение
	Fixed bool

	// путь к файлу (по умолчанию stdin)
	File string

	// шаблон поиска
	Pattern string
}

func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var in io.Reader = os.Stdin
	if opts.File != "" {
		f, err := os.Open(opts.File)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	if err := Grep(in, os.Stdout, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOptions(args []string) (Options, error) {
	var opts Options

	fs := flag.NewFlagSet("grep", flag.ContinueOnError)

	fs.IntVar(&opts.After, "A", 0, "print N lines after each match")
	fs.IntVar(&opts.Before, "B", 0, "print N lines before each match")
	fs.IntVar(&opts.Context, "C", 0, "print N lines of context around each match")
	fs.BoolVar(&opts.CountLines, "c", false, "print only the count of matching lines")
	fs.BoolVar(&opts.IgnoreCase, "i", false, "ignore case distinctions in patterns and data")
	fs.BoolVar(&opts.Invert, "v", false, "invert the sense of matching, to select non-matching lines")
	fs.BoolVar(&opts.Fixed, "F", false, "treat the pattern as a fixed string")
	fs.BoolVar(&opts.ShowLine, "n", false, "number all output lines")
	fs.StringVar(&opts.File, "file", "", "read input from file")

	if err := fs.Parse(args); err != nil {
		return opts, err
	}

	if fs.NArg() == 0 {
		return opts, errors.New("no pattern specified")
	}

	opts.Pattern = fs.Arg(0)

	return opts, nil
}
