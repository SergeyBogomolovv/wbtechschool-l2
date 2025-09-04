package sort_test

import (
	"bytes"
	sort "mysort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	testCases := []struct {
		name  string
		input []string
		want  []string
		opts  sort.Options
	}{
		{
			name: "example/basic lexicographic",
			input: []string{
				"a", "c", "b",
			},
			want: []string{
				"a", "b", "c",
			},
		},
		{
			name: "numeric ascending (-n)",
			input: []string{
				"10", "2", "-3", "2",
			},
			want: []string{
				"-3", "2", "2", "10",
			},
			opts: sort.Options{
				Numeric: true,
			},
		},
		{
			name: "numeric descending (-nr)",
			input: []string{
				"10", "2", "-3", "2",
			},
			want: []string{
				"10", "2", "2", "-3",
			},
			opts: sort.Options{
				Numeric: true,
				Reverse: true,
			},
		},
		{
			name: "unique lines (-u)",
			input: []string{
				"a", "b", "a", "a", "b", "c",
			},
			want: []string{
				"a", "b", "c",
			},
			opts: sort.Options{
				Unique: true,
			},
		},
		{
			name: "month names (-M)",
			input: []string{
				"Feb", "Jan", "Dec", "Aug",
			},
			want: []string{
				"Jan", "Feb", "Aug", "Dec",
			},
			opts: sort.Options{
				Month: true,
			},
		},
		{
			name: "column sort by 2nd field (-k 2) numeric (-n)",
			input: []string{
				"b\t2",
				"a\t10",
				"c\t1",
			},
			want: []string{
				"c\t1",
				"b\t2",
				"a\t10",
			},
			opts: sort.Options{
				Column:  2,
				Numeric: true,
			},
		},
		{
			name: "ignore trailing blanks (-b)",
			input: []string{
				"b",
				"a   ",
				"a  ",
			},
			want: []string{
				"a",
				"a",
				"b",
			},
			opts: sort.Options{
				IgnoreTrailingBlanks: true,
			},
		},
		{
			name: "ignore trailing blanks and unique",
			input: []string{
				"b",
				"a   ",
				"a  ",
			},
			want: []string{
				"a",
				"b",
			},
			opts: sort.Options{
				IgnoreTrailingBlanks: true,
				Unique:               true,
			},
		},
		{
			name: "human numeric (-h)",
			input: []string{
				"1K", "512", "2M", "1M", "10K",
			},
			want: []string{
				"512", "1K", "10K", "1M", "2M",
			},
			opts: sort.Options{
				HumanNumeric: true,
			},
		},
		{
			name: "combined: -k2 -n -r (column numeric reverse)",
			input: []string{
				"id1\t3",
				"id2\t7",
				"id3\t1",
			},
			want: []string{
				"id2\t7",
				"id1\t3",
				"id3\t1",
			},
			opts: sort.Options{
				Column:  2,
				Numeric: true,
				Reverse: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := bytes.NewBufferString(strings.Join(tc.input, "\n") + "\n")
			var out bytes.Buffer

			err := sort.Sort(in, &out, tc.opts)
			require.NoError(t, err)

			got := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSort_Check_Sorted_OK(t *testing.T) {
	input := []string{"a", "b", "c"}

	in := bytes.NewBufferString(strings.Join(input, "\n") + "\n")
	var out bytes.Buffer

	err := sort.Sort(in, &out, sort.Options{Check: true})
	assert.NoError(t, err)
}

func TestSort_Check_NotSorted_ReturnsError(t *testing.T) {
	input := []string{"b", "a", "c"}

	in := bytes.NewBufferString(strings.Join(input, "\n") + "\n")
	var out bytes.Buffer

	err := sort.Sort(in, &out, sort.Options{Check: true})
	assert.Error(t, err)
}
