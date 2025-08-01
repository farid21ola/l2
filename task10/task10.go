package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SortOptions содержит все опции сортировки
type SortOptions struct {
	keyColumn    int  // -k N
	numeric      bool // -n
	reverse      bool // -r
	unique       bool // -u
	monthSort    bool // -M
	ignoreBlanks bool // -b
	checkSorted  bool // -c
	humanNumeric bool // -h
}

// Line представляет строку для сортировки с дополнительными данными
type Line struct {
	original string
	key      string
	value    interface{}
}

// MonthMap содержит соответствие названий месяцев и их номеров
var monthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

// parseHumanNumeric парсит человекочитаемые размеры (K, M, G)
func parseHumanNumeric(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))

	var multiplier float64 = 1
	var suffix string

	if strings.HasSuffix(s, "k") {
		multiplier = 1024
		suffix = "k"
	} else if strings.HasSuffix(s, "m") {
		multiplier = 1024 * 1024
		suffix = "m"
	} else if strings.HasSuffix(s, "g") {
		multiplier = 1024 * 1024 * 1024
		suffix = "g"
	} else if strings.HasSuffix(s, "t") {
		multiplier = 1024 * 1024 * 1024 * 1024
		suffix = "t"
	}

	if suffix != "" {
		s = strings.TrimSuffix(s, suffix)
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

// parseMonth парсит название месяца
func parseMonth(s string) (int, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if month, exists := monthMap[s]; exists {
		return month, nil
	}
	return 0, fmt.Errorf("invalid month: %s", s)
}

// extractKey извлекает ключ для сортировки из строки
func extractKey(line string, opts SortOptions) string {
	if opts.keyColumn <= 0 {
		return line
	}

	// Разделяем по табуляции
	fields := strings.Split(line, "\t")
	if opts.keyColumn > len(fields) {
		return ""
	}

	key := fields[opts.keyColumn-1]
	if opts.ignoreBlanks {
		key = strings.TrimSpace(key)
	}

	return key
}

// parseValue парсит значение в зависимости от типа сортировки
func parseValue(key string, opts SortOptions) interface{} {
	if opts.numeric {
		if val, err := strconv.ParseFloat(key, 64); err == nil {
			return val
		}
		return 0.0
	}

	if opts.monthSort {
		if val, err := parseMonth(key); err == nil {
			return val
		}
		return 0
	}

	if opts.humanNumeric {
		if val, err := parseHumanNumeric(key); err == nil {
			return val
		}
		return 0.0
	}

	return key
}

// isSorted проверяет, отсортированы ли данные
func isSorted(lines []Line, opts SortOptions) bool {
	for i := 1; i < len(lines); i++ {
		if compareLines(lines[i-1], lines[i], opts) > 0 {
			return false
		}
	}
	return true
}

// compareLines сравнивает две строки для сортировки
func compareLines(a, b Line, opts SortOptions) int {
	var result int

	switch aVal := a.value.(type) {
	case float64:
		bVal := b.value.(float64)
		if aVal < bVal {
			result = -1
		} else if aVal > bVal {
			result = 1
		} else {
			result = 0
		}
	case int:
		bVal := b.value.(int)
		if aVal < bVal {
			result = -1
		} else if aVal > bVal {
			result = 1
		} else {
			result = 0
		}
	case string:
		bVal := b.value.(string)
		result = strings.Compare(aVal, bVal)
	}

	if opts.reverse {
		result = -result
	}

	return result
}

// removeDuplicates удаляет дубликаты из отсортированного списка
func removeDuplicates(lines []Line, opts SortOptions) []Line {
	if len(lines) == 0 {
		return lines
	}

	result := make([]Line, 0, len(lines))
	result = append(result, lines[0])

	for i := 1; i < len(lines); i++ {
		if compareLines(lines[i-1], lines[i], opts) != 0 {
			result = append(result, lines[i])
		}
	}
	return result
}

// readLines читает строки из файла или STDIN
func readLines(filename string) ([]string, error) {
	var file *os.File
	var err error
	var fileSize int64

	if filename == "" || filename == "-" {
		file = os.Stdin
		fileSize = 0 // Неизвестный размер для STDIN
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Получаем размер файла
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		fileSize = fileInfo.Size()
	}

	scanner := bufio.NewScanner(file)

	var maxCapacity int
	var initialCapacity int

	if fileSize == 0 {
		// STDIN или неизвестный размер - используем средние значения
		maxCapacity = 64 * 1024 // 64KB
		initialCapacity = 1000
	} else if fileSize < 1024*1024 {
		// Маленький файл (< 1MB)
		maxCapacity = 32 * 1024 // 32KB
		initialCapacity = 100
	} else if fileSize < 100*1024*1024 {
		// Средний файл (1MB - 100MB)
		maxCapacity = 64 * 1024 // 64KB
		initialCapacity = 10000
	} else {
		// Большой файл (> 100MB)
		maxCapacity = 1024 * 1024 // 1MB
		initialCapacity = 100000
	}

	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var lines []string
	lines = make([]string, 0, initialCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func main() {
	// Определяем флаги
	var (
		keyColumn    = flag.Int("k", 0, "сортировать по столбцу N")
		numeric      = flag.Bool("n", false, "сортировать по числовому значению")
		reverse      = flag.Bool("r", false, "сортировать в обратном порядке")
		unique       = flag.Bool("u", false, "не выводить повторяющиеся строки")
		monthSort    = flag.Bool("M", false, "сортировать по названию месяца")
		ignoreBlanks = flag.Bool("b", false, "игнорировать хвостовые пробелы")
		checkSorted  = flag.Bool("c", false, "проверить, отсортированы ли данные")
		humanNumeric = flag.Bool("h", false, "сортировать по человекочитаемым размерам")
	)

	flag.Parse()

	// Создаём опции сортировки
	opts := SortOptions{
		keyColumn:    *keyColumn,
		numeric:      *numeric,
		reverse:      *reverse,
		unique:       *unique,
		monthSort:    *monthSort,
		ignoreBlanks: *ignoreBlanks,
		checkSorted:  *checkSorted,
		humanNumeric: *humanNumeric,
	}

	// Получаем имя файла из аргументов
	var filename string
	if len(flag.Args()) > 0 {
		filename = flag.Args()[0]
	}

	// Читаем строки
	lines, err := readLines(filename)
	if err != nil {
		fmt.Printf("Ошибка чтения файла: %v", err)
		os.Exit(1)
	}

	sortLines := make([]Line, 0, len(lines))
	// Преобразуем строки в структуры для сортировки
	for _, line := range lines {
		key := extractKey(line, opts)
		value := parseValue(key, opts)
		sortLines = append(sortLines, Line{
			original: line,
			key:      key,
			value:    value,
		})
	}

	// Проверяем, отсортированы ли данные
	if opts.checkSorted {
		if isSorted(sortLines, opts) {
			fmt.Println("Данные отсортированы")
		} else {
			fmt.Println("Данные не отсортированы")
		}
		return
	}

	// Сортируем строки
	sort.Slice(sortLines, func(i, j int) bool {
		return compareLines(sortLines[i], sortLines[j], opts) < 0
	})

	// Удаляем дубликаты если нужно
	if opts.unique {
		sortLines = removeDuplicates(sortLines, opts)
	}

	// Выводим результат сортировки
	for _, line := range sortLines {
		fmt.Println(line.original)
	}
}
