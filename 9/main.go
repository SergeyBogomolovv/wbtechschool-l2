package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	fmt.Println(DecodeString("a4bc2d5e")) // "aaaabccddddde"
	fmt.Println(DecodeString("abcd"))     // "abcd"
	fmt.Println(DecodeString("45"))       // "err"
}

var (
	// ErrInvalidString is returned when string is invalid.
	ErrInvalidString = errors.New("invalid string")
)

// DecodeString returns decoded string. Example: "a4bc2d5e" -> "aaaabccddddde".
func DecodeString(s string) (string, error) {
	runes := []rune(s)
	var prev rune
	var result strings.Builder
	isPrevLetter := false

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if !unicode.IsDigit(ch) {
			if isPrevLetter {
				result.WriteRune(prev)
			}
			if ch == '\\' {
				i++
				if i == len(runes) {
					return "", ErrInvalidString
				}
				ch = runes[i]
			}
			prev = ch
			isPrevLetter = true
			continue
		}

		if !isPrevLetter {
			return "", ErrInvalidString
		}

		j := i
		for j < len(runes) && unicode.IsDigit(runes[j]) {
			j++
		}
		num, err := strconv.Atoi(string(runes[i:j]))
		if err != nil {
			return "", fmt.Errorf("failed to parse number: %w", err)
		}
		i = j - 1

		for range num {
			result.WriteRune(prev)
		}
		isPrevLetter = false
	}

	if isPrevLetter {
		result.WriteRune(prev)
	}

	return result.String(), nil
}
