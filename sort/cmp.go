package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Options - sort options.
type Options struct {
	Column               int
	Numeric              bool
	Month                bool
	Reverse              bool
	IgnoreTrailingBlanks bool
	HumanNumeric         bool
}

var (
	months = map[string]int{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}
)

// конструктор функции сравнения
func cmpFunc(opts Options) func(a, b string) int {
	key := func(s string) string {
		if opts.IgnoreTrailingBlanks {
			s = strings.TrimRight(s, " \t")
		}
		if opts.Column > 0 {
			fields := strings.Fields(s)
			if opts.Column-1 < len(fields) {
				return fields[opts.Column-1]
			}
			return ""
		}
		return s
	}

	return func(a, b string) int {
		a, b = key(a), key(b)
		var cmp int

		switch {

		case opts.Month:
			if len(a) < 3 {
				fmt.Fprintf(os.Stderr, "invalid month value: %s\n", a)
				os.Exit(1)
			}
			if len(b) < 3 {
				fmt.Fprintf(os.Stderr, "invalid month value: %s\n", b)
				os.Exit(1)
			}
			monthA, monthB := months[strings.ToLower(a[:3])], months[strings.ToLower(b[:3])]
			cmp = monthA - monthB

		case opts.Numeric:
			fa, err := strconv.ParseFloat(a, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid numeric value: %s\n", a)
				os.Exit(1)
			}
			fb, err := strconv.ParseFloat(b, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid numeric value: %s\n", b)
				os.Exit(1)
			}

			switch {
			case fa < fb:
				cmp = -1
			case fa > fb:
				cmp = 1
			default:
				cmp = 0
			}

		case opts.HumanNumeric:
			cmp = humanNumericCompare(a, b)

		default:
			cmp = strings.Compare(a, b)
		}

		if opts.Reverse {
			cmp = -cmp
		}

		return cmp
	}
}

func parseHumanSize(s string) float64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0
	}

	numPart := s
	suffix := ""

	for i := len(s) - 1; i >= 0; i-- {
		if unicode.IsDigit(rune(s[i])) || s[i] == '.' {
			numPart = s[:i+1]
			suffix = s[i+1:]
			break
		}
	}

	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid value: %s\n", s)
		os.Exit(1)
	}

	mult := map[string]float64{
		"":  1,
		"B": 1,
		"K": math.Pow(1024, 1),
		"M": math.Pow(1024, 2),
		"G": math.Pow(1024, 3),
		"T": math.Pow(1024, 4),
		"P": math.Pow(1024, 5),
		"E": math.Pow(1024, 6),
	}

	m := mult[suffix]
	if m == 0 {
		m = 1
	}

	return val * m
}

func humanNumericCompare(a, b string) int {
	fa := parseHumanSize(a)
	fb := parseHumanSize(b)

	switch {
	case fa < fb:
		return -1
	case fa > fb:
		return 1
	default:
		return 0
	}
}
