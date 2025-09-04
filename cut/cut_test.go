package cut_test

import (
	"bytes"
	"cut"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCut(t *testing.T) {
	testCases := []struct {
		name  string
		input []string
		want  []string
		opts  cut.Options
	}{
		{
			name:  "example",
			input: []string{"test;data;test", "ignored"},
			opts: cut.Options{
				Fields:    []int{1, 3},
				Separated: true,
				Delimiter: ";",
			},
			want: []string{"test;test"},
		},
		{
			name:  "default_delimiter_tab",
			input: []string{"a\tb\tc", "x\ty\tz"},
			opts: cut.Options{
				Fields: []int{2},
			},
			want: []string{"b", "y"},
		},
		{
			name:  "custom_delimiter_comma",
			input: []string{"a,b,c", "1,2,3"},
			opts: cut.Options{
				Fields:    []int{1, 3},
				Delimiter: ",",
			},
			want: []string{"a,c", "1,3"},
		},
		{
			name:  "separated_true_skips_lines_without_delimiter",
			input: []string{"no_delims_here", "a;b;c", "also no delim"},
			opts: cut.Options{
				Fields:    []int{2},
				Separated: true,
				Delimiter: ";",
			},
			want: []string{"b"},
		},
		{
			name:  "separated_false_prints_lines_without_delimiter_unchanged",
			input: []string{"no_delims_here", "a;b;c"},
			opts: cut.Options{
				Fields:    []int{2},
				Separated: false,
				Delimiter: ";",
			},
			want: []string{"no_delims_here", "b"},
		},
		{
			name:  "out_of_bounds_fields_are_ignored",
			input: []string{"a;b", "1;2"},
			opts: cut.Options{
				Fields:    []int{2, 3, 9},
				Delimiter: ";",
			},
			want: []string{"b", "2"},
		},
		{
			name:  "multiple_fields_unsorted_and_preserve_requested_order",
			input: []string{"a;b;c;d"},
			opts: cut.Options{
				Fields:    []int{3, 1, 4},
				Delimiter: ";",
			},
			want: []string{"c;a;d"},
		},
		{
			name:  "consecutive_delimiters_empty_fields_retained",
			input: []string{"a;;c;", ";b;;"},
			opts: cut.Options{
				Fields:    []int{1, 2, 3, 4},
				Delimiter: ";",
			},
			want: []string{"a;;c;", ";b;;"},
		},
		{
			name:  "trailing_and_leading_delimiter_create_empty_fields",
			input: []string{";a;b;", ";x;"},
			opts: cut.Options{
				Fields:    []int{1, 2, 3, 4},
				Delimiter: ";",
			},
			want: []string{";a;b;", ";x;"},
		},
		{
			name:  "single_field_first",
			input: []string{"a;b;c", "x;y;z"},
			opts: cut.Options{
				Fields:    []int{1},
				Delimiter: ";",
			},
			want: []string{"a", "x"},
		},
		{
			name:  "single_field_last",
			input: []string{"a;b;c", "x;y;z"},
			opts: cut.Options{
				Fields:    []int{3},
				Delimiter: ";",
			},
			want: []string{"c", "z"},
		},
		{
			name:  "range_like_usage_by_expansion_1_3_5",
			input: []string{"f1;f2;f3;f4;f5"},
			opts: cut.Options{
				Fields:    []int{1, 3, 4, 5},
				Delimiter: ";",
			},
			want: []string{"f1;f3;f4;f5"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := bytes.NewBufferString(strings.Join(tc.input, "\n") + "\n")
			var out bytes.Buffer

			err := cut.Cut(in, &out, tc.opts)
			require.NoError(t, err)
			result := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name string
		spec string
		want []int
	}{
		{
			name: "single_field",
			spec: "1",
			want: []int{1},
		},
		{
			name: "simple_list",
			spec: "1,3,5",
			want: []int{1, 3, 5},
		},
		{
			name: "range_inclusive",
			spec: "3-5",
			want: []int{3, 4, 5},
		},
		{
			name: "mixed_list_and_range",
			spec: "1,3-5,9",
			want: []int{1, 3, 4, 5, 9},
		},
		{
			name: "duplicates_collapsed_preserve_first_order",
			spec: "1,1,2,2,3",
			want: []int{1, 2, 3},
		},
		{
			name: "overlapping_ranges_dedup_preserve_first_order",
			spec: "1-3,2-4",
			want: []int{1, 2, 3, 4},
		},
		{
			name: "large_numbers_ok",
			spec: "1,1000-1002",
			want: []int{1, 1000, 1001, 1002},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cut.ParseFields(tt.spec)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
