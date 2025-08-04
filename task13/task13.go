package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CutOptions содержит все опции для утилиты cut
type CutOptions struct {
	fields    string // -f
	delimiter string // -d
	separated bool   // -s
}

// FieldRange представляет диапазон полей
type FieldRange struct {
	start int
	end   int
}

// parseFieldRanges парсит строку с номерами полей и диапазонами
func parseFieldRanges(fieldsStr string) ([]FieldRange, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("пустая строка полей")
	}

	var ranges []FieldRange
	parts := strings.Split(fieldsStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("неверный формат диапазона: %s", part)
			}

			startStr := strings.TrimSpace(rangeParts[0])
			endStr := strings.TrimSpace(rangeParts[1])

			var start, end int
			var err error

			if startStr == "" {
				start = 1
			} else {
				start, err = strconv.Atoi(startStr)
				if err != nil {
					return nil, fmt.Errorf("неверный номер поля: %s", startStr)
				}
			}

			if endStr == "" {
				end = -1
			} else {
				end, err = strconv.Atoi(endStr)
				if err != nil {
					return nil, fmt.Errorf("неверный номер поля: %s", endStr)
				}
			}

			if start > 0 && end > 0 && start > end {
				return nil, fmt.Errorf("неверный диапазон: %d-%d", start, end)
			}

			ranges = append(ranges, FieldRange{start: start, end: end})
		} else {
			fieldNum, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("неверный номер поля: %s", part)
			}
			if fieldNum <= 0 {
				return nil, fmt.Errorf("номер поля должен быть положительным: %d", fieldNum)
			}
			ranges = append(ranges, FieldRange{start: fieldNum, end: fieldNum})
		}
	}

	return ranges, nil
}

// shouldIncludeField проверяет, должно ли поле быть включено в вывод
func shouldIncludeField(fieldNum int, ranges []FieldRange) bool {
	for _, r := range ranges {
		if r.start <= fieldNum && (r.end == -1 || fieldNum <= r.end) {
			return true
		}
	}
	return false
}

// processLine обрабатывает одну строку согласно опциям cut
func processLine(line string, opts CutOptions, ranges []FieldRange) (string, bool) {
	if opts.separated && !strings.Contains(line, opts.delimiter) {
		return "", false
	}

	fields := strings.Split(line, opts.delimiter)

	if len(ranges) == 0 {
		return "", true
	}

	var result []string

	for i, field := range fields {
		fieldNum := i + 1 // Номера полей начинаются с 1
		if shouldIncludeField(fieldNum, ranges) {
			result = append(result, field)
		}
	}

	return strings.Join(result, opts.delimiter), true
}

// readLines читает строки из STDIN
func readLines() ([]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	const maxCapacity = 64 * 1024 // 64KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// printResult обрабатывает и выводит результат для всех строк
func printResult(lines []string, opts CutOptions, ranges []FieldRange) {
	for _, line := range lines {
		result, shouldOutput := processLine(line, opts, ranges)
		if shouldOutput {
			fmt.Println(result)
		}
	}
}

func main() {
	// Определяем флаги
	var (
		fields    = flag.String("f", "", "номера полей для вывода (например: 1,3-5)")
		delimiter = flag.String("d", "\t", "разделитель полей")
		separated = flag.Bool("s", false, "только строки, содержащие разделитель")
	)

	flag.Parse()

	// Создаём опции
	opts := CutOptions{
		fields:    *fields,
		delimiter: *delimiter,
		separated: *separated,
	}

	// Проверяем, что указаны поля для вывода
	if opts.fields == "" {
		fmt.Fprintf(os.Stderr, "Ошибка: необходимо указать поля для вывода (-f)\n")
		os.Exit(1)
	}

	// Проверяем, что нет лишних аргументов
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Ошибка: программа читает данные только из STDIN\n")
		os.Exit(1)
	}

	// Парсим диапазоны полей
	ranges, err := parseFieldRanges(opts.fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга полей: %v\n", err)
		os.Exit(1)
	}

	// Читаем строки из STDIN
	lines, err := readLines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения входных данных: %v\n", err)
		os.Exit(1)
	}

	// Обрабатываем каждую строку
	printResult(lines, opts, ranges)
}
