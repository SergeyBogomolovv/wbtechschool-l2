package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSort(t *testing.T) {
	testCases := []struct {
		name     string
		opts     SortOptions
		data     []string
		expected []string
	}{
		{
			name:     "lexicographic",
			opts:     SortOptions{},
			data:     []string{"c", "a", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "numeric",
			opts:     SortOptions{Numeric: true},
			data:     []string{"10", "2", "1"},
			expected: []string{"1", "2", "10"},
		},
		{
			name:     "reverse",
			opts:     SortOptions{Reverse: true},
			data:     []string{"a", "b", "c"},
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "human readable",
			opts:     SortOptions{HumanNumeric: true},
			data:     []string{"1K", "900", "2K"},
			expected: []string{"900", "1K", "2K"},
		},
		{
			name:     "month",
			opts:     SortOptions{Month: true},
			data:     []string{"Mar", "Jan", "Feb"},
			expected: []string{"Jan", "Feb", "Mar"},
		},
		{
			name: "column sort second column",
			opts: SortOptions{Column: 2},
			data: []string{
				"z\tb",
				"x\ta",
				"y\tc",
			},
			expected: []string{
				"x\ta",
				"z\tb",
				"y\tc",
			},
		},
		{
			name: "column numeric sort second column",
			opts: SortOptions{Column: 2, Numeric: true},
			data: []string{
				"z\t10",
				"x\t2",
				"y\t1",
			},
			expected: []string{
				"y\t1",
				"x\t2",
				"z\t10",
			},
		},
		{
			name:     "ignore trailing blanks",
			opts:     SortOptions{IgnoreTrailingBlanks: true},
			data:     []string{"abc   ", "abd", "abc"},
			expected: []string{"abc   ", "abc", "abd"},
		},
		{
			name:     "reverse numeric",
			opts:     SortOptions{Numeric: true, Reverse: true},
			data:     []string{"1", "2", "10"},
			expected: []string{"10", "2", "1"},
		},
		{
			name:     "human numeric with suffixes",
			opts:     SortOptions{HumanNumeric: true},
			data:     []string{"1M", "512K", "2M"},
			expected: []string{"512K", "1M", "2M"},
		},
		{
			name:     "month mixed case",
			opts:     SortOptions{Month: true},
			data:     []string{"dec", "Aug", "JAN"},
			expected: []string{"JAN", "Aug", "dec"},
		},
		{
			name:     "unknown month fallback to lexicographic",
			opts:     SortOptions{Month: true},
			data:     []string{"abc", "Jan", "zzz"},
			expected: []string{"Jan", "abc", "zzz"},
		},
		{
			name:     "empty lines",
			opts:     SortOptions{},
			data:     []string{"c", "", "a"},
			expected: []string{"", "a", "c"},
		},
		{
			name:     "numeric invalid fallback to string compare",
			opts:     SortOptions{Numeric: true},
			data:     []string{"10", "abc", "2"},
			expected: []string{"2", "10", "abc"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := bytes.NewBufferString(strings.Join(tc.data, "\n") + "\n")
			var out bytes.Buffer

			if err := ExternalSort(in, &out, cmpFunc(tc.opts)); err != nil {
				t.Fatalf("ExternalSort failed: %v", err)
			}

			result := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			if len(result) != len(tc.expected) {
				t.Fatalf("wrong number of lines: got %d, want %d", len(result), len(tc.expected))
			}
			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("line %d: got %q, want %q", i, result[i], tc.expected[i])
				}
			}
		})
	}
}
