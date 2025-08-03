package main

import (
	"reflect"
	"sort"
	"testing"
)

// TestExtractKey тестирует функцию извлечения ключа
func TestExtractKey(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		opts     SortOptions
		expected string
	}{
		{
			name:     "без указания столбца",
			line:     "apple\t5\tJan",
			opts:     SortOptions{keyColumn: 0},
			expected: "apple\t5\tJan",
		},
		{
			name:     "первый столбец",
			line:     "apple\t5\tJan",
			opts:     SortOptions{keyColumn: 1},
			expected: "apple",
		},
		{
			name:     "столбец с пробелами",
			line:     "apple\t 5 \tJan",
			opts:     SortOptions{keyColumn: 2, ignoreBlanks: true},
			expected: "5",
		},
		{
			name:     "несуществующий столбец",
			line:     "apple\t5",
			opts:     SortOptions{keyColumn: 5},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractKey(test.line, test.opts)
			if result != test.expected {
				t.Errorf("ожидалось %q, получено %q", test.expected, result)
			}
		})
	}
}

// TestParseValue тестирует функцию парсинга значений
func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		opts     SortOptions
		expected interface{}
	}{
		{
			name:     "обычная строка",
			key:      "apple",
			opts:     SortOptions{},
			expected: "apple",
		},
		{
			name:     "числовое значение",
			key:      "123",
			opts:     SortOptions{numeric: true},
			expected: 123.0,
		},
		{
			name:     "нечисловое значение при числовой сортировке",
			key:      "abc",
			opts:     SortOptions{numeric: true},
			expected: 0.0,
		},
		{
			name:     "месяц Dec",
			key:      "Dec",
			opts:     SortOptions{monthSort: true},
			expected: 12,
		},
		{
			name:     "неверный месяц",
			key:      "Invalid",
			opts:     SortOptions{monthSort: true},
			expected: 0,
		},
		{
			name:     "человекочитаемый размер 1K",
			key:      "1K",
			opts:     SortOptions{humanNumeric: true},
			expected: 1024.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseValue(test.key, test.opts)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("ожидалось %v, получено %v", test.expected, result)
			}
		})
	}
}

// TestSortLinesIntegration тестирует полную интеграцию сортировки
func TestSortLinesIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		opts     SortOptions
		expected []string
	}{
		{
			name:     "базовая сортировка",
			input:    []string{"cherry", "banana", "apple"},
			opts:     SortOptions{},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "числовая сортировка",
			input:    []string{"10", "2", "1", "20"},
			opts:     SortOptions{numeric: true},
			expected: []string{"1", "2", "10", "20"},
		},
		{
			name:     "обратная сортировка",
			input:    []string{"apple", "banana", "cherry"},
			opts:     SortOptions{reverse: true},
			expected: []string{"cherry", "banana", "apple"},
		},
		{
			name:     "с дубликатами",
			input:    []string{"apple", "banana", "apple", "cherry"},
			opts:     SortOptions{unique: true},
			expected: []string{"apple", "banana", "cherry"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Создаём структуры Line
			lines := make([]Line, len(test.input))
			for i, line := range test.input {
				key := extractKey(line, test.opts)
				value := parseValue(key, test.opts)
				lines[i] = Line{
					original: line,
					key:      key,
					value:    value,
				}
			}

			// Сортируем
			sort.Slice(lines, func(i, j int) bool {
				return compareLines(lines[i], lines[j], test.opts) < 0
			})

			// Удаляем дубликаты если нужно
			if test.opts.unique {
				lines = removeDuplicates(lines, test.opts)
			}

			// Проверяем результат
			result := make([]string, len(lines))
			for i, line := range lines {
				result[i] = line.original
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("ожидалось %v, получено %v", test.expected, result)
			}
		})
	}
}
