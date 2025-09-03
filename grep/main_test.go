package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrep(t *testing.T) {
	input := []string{
		"Alpha",
		"bravo",
		"CHARLIE",
		"delta match here",
		"echo",
		"match again",
		"Foxtrot",
		"golf",
		"hotel MATCH",
		"india",
	}

	testCases := []struct {
		name     string
		opts     Options
		input    []string
		expected []string
	}{
		{
			name:  "basic regex match",
			opts:  Options{Pattern: "match"},
			input: input,
			expected: []string{
				"delta match here",
				"match again",
			},
		},
		{
			name:  "fixed substring match -F",
			opts:  Options{Pattern: "MATCH", Fixed: true},
			input: input,
			expected: []string{
				"hotel MATCH",
			},
		},
		{
			name:  "ignore case -i with regex",
			opts:  Options{Pattern: "match", IgnoreCase: true},
			input: input,
			expected: []string{
				"delta match here",
				"match again",
				"hotel MATCH",
			},
		},
		{
			name:  "ignore case with fixed -F -i",
			opts:  Options{Pattern: "charlie", Fixed: true, IgnoreCase: true},
			input: input,
			expected: []string{
				"CHARLIE",
			},
		},
		{
			name:  "invert -v (non-matching lines only)",
			opts:  Options{Pattern: "match", Invert: true},
			input: input[:6],
			expected: []string{
				"Alpha",
				"bravo",
				"CHARLIE",
				"echo",
			},
		},
		{
			name:     "count only -c ignores context",
			opts:     Options{Pattern: "match", CountLines: true, After: 2, Before: 2},
			input:    input,
			expected: []string{"2"},
		},
		{
			name:  "line numbers -n",
			opts:  Options{Pattern: "match", ShowLine: true},
			input: input,
			expected: []string{
				"4:delta match here",
				"6:match again",
			},
		},
		{
			name:  "after context -A 2 with numbering",
			opts:  Options{Pattern: "CHARLIE", After: 2, ShowLine: true, Fixed: true},
			input: input,
			expected: []string{
				"3:CHARLIE",
				"4:delta match here",
				"5:echo",
			},
		},
		{
			name:  "before context -B 2 at start boundary (no duplicates)",
			opts:  Options{Pattern: "^A", Before: 2, ShowLine: true},
			input: input,
			expected: []string{
				"1:Alpha",
			},
		},
		{
			name:  "around context -C 1 equals -A1 -B1 with overlaps",
			opts:  Options{Pattern: "match", Context: 1, ShowLine: true},
			input: input,
			expected: []string{
				"3:CHARLIE",
				"4:delta match here",
				"5:echo",
				"6:match again",
				"7:Foxtrot",
			},
		},
		{
			name:  "around context -C 1 without -n to test formatting",
			opts:  Options{Pattern: "bravo|Foxtrot", Context: 1},
			input: input,
			expected: []string{
				"Alpha",
				"bravo",
				"CHARLIE",
				"match again",
				"Foxtrot",
				"golf",
			},
		},
		{
			name:  "multiple adjacent matches produce continuous context once",
			opts:  Options{Pattern: "golf|hotel", Context: 1, ShowLine: true},
			input: input,
			expected: []string{
				"7:Foxtrot",
				"8:golf",
				"9:hotel MATCH",
				"10:india",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := bytes.NewBufferString(strings.Join(tc.input, "\n") + "\n")
			var out bytes.Buffer
			err := Grep(in, &out, tc.opts)
			require.NoError(t, err)

			result := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			require.Equal(t, tc.expected, result)
		})
	}
}
