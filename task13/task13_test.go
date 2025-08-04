package main

import (
	"testing"
)

// TestParseFieldRanges проверяет парсинг диапазонов полей
func TestParseFieldRanges(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []FieldRange
		expectError bool
	}{
		{
			name:     "одно поле",
			input:    "1",
			expected: []FieldRange{{start: 1, end: 1}},
		},
		{
			name:     "несколько полей",
			input:    "1,3",
			expected: []FieldRange{{start: 1, end: 1}, {start: 3, end: 3}},
		},
		{
			name:     "диапазон полей",
			input:    "2-4",
			expected: []FieldRange{{start: 2, end: 4}},
		},
		{
			name:     "диапазон от начала",
			input:    "-3",
			expected: []FieldRange{{start: 1, end: 3}},
		},
		{
			name:     "диапазон до конца",
			input:    "2-",
			expected: []FieldRange{{start: 2, end: -1}},
		},
		{
			name:     "комбинированные поля и диапазоны",
			input:    "1,3-4",
			expected: []FieldRange{{start: 1, end: 1}, {start: 3, end: 4}},
		},
		{
			name:        "пустая строка",
			input:       "",
			expectError: true,
		},
		{
			name:        "неверный формат диапазона",
			input:       "1-2-3",
			expectError: true,
		},
		{
			name:        "отрицательный номер поля",
			input:       "0",
			expectError: true,
		},
		{
			name:        "неверный номер поля",
			input:       "abc",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseFieldRanges(test.input)

			if test.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка, но её не было")
				}
			} else {
				if err != nil {
					t.Errorf("неожиданная ошибка: %v", err)
				}
				if !compareFieldRanges(result, test.expected) {
					t.Errorf("ожидалось %v, получено %v", test.expected, result)
				}
			}
		})
	}
}

// TestProcessLineWithCustomDelimiter проверяет обработку с пользовательским разделителем
func TestProcessLineWithCustomDelimiter(t *testing.T) {
	opts := CutOptions{
		fields:    "1,3",
		delimiter: ",",
		separated: false,
	}
	ranges := []FieldRange{{start: 1, end: 1}, {start: 3, end: 3}}

	tests := []struct {
		name           string
		line           string
		expectedResult string
		expectedOutput bool
	}{
		{
			name:           "строка с запятыми",
			line:           "apple,red,fruit,sweet",
			expectedResult: "apple,fruit",
			expectedOutput: true,
		},
		{
			name:           "строка с табуляцией",
			line:           "apple\tred\tfruit",
			expectedResult: "apple\tred\tfruit",
			expectedOutput: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, shouldOutput := processLine(test.line, opts, ranges)
			if result != test.expectedResult {
				t.Errorf("ожидался результат %q, получено %q", test.expectedResult, result)
			}
			if shouldOutput != test.expectedOutput {
				t.Errorf("ожидался вывод %v, получено %v", test.expectedOutput, shouldOutput)
			}
		})
	}
}

// TestProcessLineWithRanges проверяет обработку с различными диапазонами
func TestProcessLineWithRanges(t *testing.T) {
	opts := CutOptions{
		fields:    "2-4",
		delimiter: "\t",
		separated: false,
	}
	ranges := []FieldRange{{start: 2, end: 4}}

	tests := []struct {
		name           string
		line           string
		expectedResult string
		expectedOutput bool
	}{
		{
			name:           "диапазон полей",
			line:           "apple\tred\tfruit\tsweet",
			expectedResult: "red\tfruit\tsweet",
			expectedOutput: true,
		},
		{
			name:           "меньше полей чем диапазон",
			line:           "apple\tred",
			expectedResult: "red",
			expectedOutput: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, shouldOutput := processLine(test.line, opts, ranges)
			if result != test.expectedResult {
				t.Errorf("ожидался результат %q, получено %q", test.expectedResult, result)
			}
			if shouldOutput != test.expectedOutput {
				t.Errorf("ожидался вывод %v, получено %v", test.expectedOutput, shouldOutput)
			}
		})
	}
}

// TestPrintResult проверяет обработку множества строк
func TestPrintResult(t *testing.T) {
	opts := CutOptions{
		fields:    "1,3",
		delimiter: "\t",
		separated: false,
	}
	ranges := []FieldRange{{start: 1, end: 1}, {start: 3, end: 3}}

	lines := []string{
		"apple\tred\tfruit\tsweet",
		"banana\tyellow\tfruit\tsweet",
		"carrot\torange\tvegetable",
	}

	// Обрабатываем каждую строку и собираем результат
	var results []string
	for _, line := range lines {
		result, shouldOutput := processLine(line, opts, ranges)
		if shouldOutput {
			results = append(results, result)
		}
	}

	expected := []string{
		"apple\tfruit",
		"banana\tfruit",
		"carrot\tvegetable",
	}

	if len(results) != len(expected) {
		t.Errorf("ожидалось %d строк, получено %d", len(expected), len(results))
		return
	}

	for i, result := range results {
		if result != expected[i] {
			t.Errorf("строка %d: ожидалось %q, получено %q", i+1, expected[i], result)
		}
	}
}

// compareFieldRanges сравнивает два слайса FieldRange
func compareFieldRanges(a, b []FieldRange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].start != b[i].start || a[i].end != b[i].end {
			return false
		}
	}
	return true
}
