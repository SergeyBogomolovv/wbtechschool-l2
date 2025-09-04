package cut

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

// Options - опции для команды cut
type Options struct {
	// Указание номеров полей, которые нужно вывести
	Fields []int
	// Использовать другой разделитель, по умолчанию табуляция
	Delimiter string
	// Только строки, содержащие разделитель
	Separated bool
}

// Cut - утилита для вырезки полей
func Cut(in io.Reader, out io.Writer, opts Options) error {
	if opts.Delimiter == "" {
		opts.Delimiter = "\t"
	}

	scanner := bufio.NewScanner(in)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, opts.Delimiter)

		if len(parts) == 1 {
			if !opts.Separated {
				fmt.Fprintln(writer, line)
			}
			continue
		}

		result := make([]string, 0, len(opts.Fields))

		for _, field := range opts.Fields {
			if field > 0 && field <= len(parts) {
				result = append(result, parts[field-1])
			}
		}

		fmt.Fprintln(writer, strings.Join(result, opts.Delimiter))
	}

	return scanner.Err()
}

// ParseFields парсит строку вида "1,3-5,7" в список уникальных номеров
func ParseFields(spec string) ([]int, error) {
	result := make(map[int]struct{})

	for part := range strings.SplitSeq(spec, ",") {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("некорректный диапазон: %s", part)
			}

			start, err1 := strconv.Atoi(bounds[0])
			end, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("некорректное число в диапазоне: %s", part)
			}

			if start > end {
				start, end = end, start
			}

			for i := start; i <= end; i++ {
				result[i] = struct{}{}
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("некорректное число: %s", part)
			}
			result[num] = struct{}{}
		}
	}

	keys := make([]int, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys, nil
}
