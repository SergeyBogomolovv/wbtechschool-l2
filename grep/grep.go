package grep

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
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

// Grep - поиск строк в файле
func Grep(in io.Reader, out io.Writer, opts Options) error {
	scanner := bufio.NewScanner(in)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// инициализируем регулярку и паттерн
	var re *regexp.Regexp
	pattern := opts.Pattern
	if !opts.Fixed {
		if opts.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}
	} else if opts.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}

	// функция для проверки подходит ли строка
	match := func(line string) bool {
		var ok bool

		if opts.Fixed {
			if opts.IgnoreCase {
				line = strings.ToLower(line)
			}
			ok = strings.Contains(line, pattern)
		} else {
			ok = re.MatchString(line)
		}

		if opts.Invert {
			ok = !ok
		}

		return ok
	}

	// если надо только посчитать вхождения, контекст не нужен
	if opts.CountLines {
		count := 0
		for scanner.Scan() {
			if match(scanner.Text()) {
				count++
			}
		}
		fmt.Fprintln(writer, count)
		return scanner.Err()
	}

	before := max(opts.Before, opts.Context)
	after := max(opts.After, opts.Context)
	pr := newContextPrinter(writer, before, after, opts.ShowLine)

	// выводим вхождения с контекстом
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		pr.handle(i, line, match(line))
	}

	return scanner.Err()
}
