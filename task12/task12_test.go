package main

import (
	"strings"
	"testing"
)

// TestGetContext тестирует функцию получения контекста
func TestGetContext(t *testing.T) {
	allLines := []Line{
		{number: 1, text: "строка 1"},
		{number: 2, text: "строка 2"},
		{number: 3, text: "строка 3"},
		{number: 4, text: "строка 4"},
		{number: 5, text: "строка 5"},
	}

	tests := []struct {
		name       string
		matchIndex int
		before     int
		after      int
		expected   []Line
	}{
		{
			name:       "контекст в середине",
			matchIndex: 2,
			before:     1,
			after:      1,
			expected: []Line{
				{number: 2, text: "строка 2"},
				{number: 4, text: "строка 4"},
			},
		},
		{
			name:       "контекст в начале",
			matchIndex: 0,
			before:     1,
			after:      1,
			expected: []Line{
				{number: 2, text: "строка 2"},
			},
		},
		{
			name:       "контекст в конце",
			matchIndex: 4,
			before:     1,
			after:      1,
			expected: []Line{
				{number: 4, text: "строка 4"},
			},
		},
		{
			name:       "большой контекст",
			matchIndex: 2,
			before:     10,
			after:      10,
			expected: []Line{
				{number: 1, text: "строка 1"},
				{number: 2, text: "строка 2"},
				{number: 4, text: "строка 4"},
				{number: 5, text: "строка 5"},
			},
		},
		{
			name:       "без контекста",
			matchIndex: 2,
			before:     0,
			after:      0,
			expected:   []Line{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getContext(allLines, test.matchIndex, test.before, test.after)
			if len(result) != len(test.expected) {
				t.Errorf("ожидалось %d элементов, получено %d", len(test.expected), len(result))
				return
			}
			for i := range result {
				if result[i].number != test.expected[i].number || result[i].text != test.expected[i].text {
					t.Errorf("элемент %d: ожидалось %+v, получено %+v", i, test.expected[i], result[i])
				}
			}
		})
	}
}

// TestGrepIntegration тестирует полную интеграцию grep
func TestGrepIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     *GrepOptions
		expected []MatchResult
	}{
		{
			name:  "базовый поиск",
			input: "Первая строка\nВторая строка с hello\nТретья строка\n",
			opts: &GrepOptions{
				pattern: "hello",
			},
			expected: []MatchResult{
				{
					line:    Line{number: 2, text: "Вторая строка с hello"},
					isMatch: true,
					context: []Line{},
				},
			},
		},
		{
			name:  "поиск с контекстом",
			input: "Первая строка\nВторая строка\nТретья строка с hello\nЧетвертая строка\n",
			opts: &GrepOptions{
				pattern:       "hello",
				beforeContext: 1,
				afterContext:  1,
			},
			expected: []MatchResult{
				{
					line:    Line{number: 3, text: "Третья строка с hello"},
					isMatch: true,
					context: []Line{
						{number: 2, text: "Вторая строка"},
						{number: 4, text: "Четвертая строка"},
					},
				},
			},
		},
		{
			name:  "инвертированный поиск",
			input: "Первая строка\nВторая строка с hello\nТретья строка\n",
			opts: &GrepOptions{
				pattern:     "hello",
				invertMatch: true,
			},
			expected: []MatchResult{
				{
					line:    Line{number: 1, text: "Первая строка"},
					isMatch: true,
					context: []Line{},
				},
				{
					line:    Line{number: 3, text: "Третья строка"},
					isMatch: true,
					context: []Line{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.input)
			matches, err := grep(reader, test.opts)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if len(matches) != len(test.expected) {
				t.Errorf("ожидалось %d совпадений, получено %d", len(test.expected), len(matches))
				return
			}

			for i, match := range matches {
				expected := test.expected[i]
				if match.line.number != expected.line.number {
					t.Errorf("совпадение %d: ожидался номер строки %d, получен %d",
						i, expected.line.number, match.line.number)
				}
				if match.line.text != expected.line.text {
					t.Errorf("совпадение %d: ожидался текст %q, получен %q",
						i, expected.line.text, match.line.text)
				}
				if match.isMatch != expected.isMatch {
					t.Errorf("совпадение %d: ожидался флаг совпадения %v, получен %v",
						i, expected.isMatch, match.isMatch)
				}
			}
		})
	}
}
