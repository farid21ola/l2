package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestSortBasic проверяет базовую сортировку
func TestSortBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{
			name:     "базовая сортировка строк",
			input:    "cherry\nbanana\napple\n",
			args:     []string{},
			expected: "apple\nbanana\ncherry\n",
		},
		{
			name:     "числовая сортировка",
			input:    "10\n2\n1\n20\n",
			args:     []string{"-n"},
			expected: "1\n2\n10\n20\n",
		},
		{
			name:     "обратная сортировка",
			input:    "apple\nbanana\ncherry\n",
			args:     []string{"-r"},
			expected: "cherry\nbanana\napple\n",
		},
		{
			name:     "удаление дубликатов",
			input:    "apple\nbanana\napple\ncherry\nbanana\n",
			args:     []string{"-u"},
			expected: "apple\nbanana\ncherry\n",
		},
		{
			name:     "сортировка по столбцу",
			input:    "cherry\t8\tFeb\nbanana\t3\tMar\napple\t5\tJan\n",
			args:     []string{"-k", "1"},
			expected: "apple\t5\tJan\nbanana\t3\tMar\ncherry\t8\tFeb\n",
		},
		{
			name:     "сортировка по месяцам",
			input:    "cherry\t8\tFeb\nbanana\t3\tMar\napple\t5\tJan\n",
			args:     []string{"-k", "3", "-M"},
			expected: "apple\t5\tJan\ncherry\t8\tFeb\nbanana\t3\tMar\n",
		},
		{
			name:     "сортировка человекочитаемых размеров",
			input:    "2M\n1K\n512K\n1G\n",
			args:     []string{"-h"},
			expected: "1K\n512K\n2M\n1G\n",
		},
		{
			name:     "числовая сортировка по столбцу",
			input:    "cherry\t8\tFeb\nbanana\t3\tMar\napple\t5\tJan\n",
			args:     []string{"-k", "2", "-n"},
			expected: "banana\t3\tMar\napple\t5\tJan\ncherry\t8\tFeb\n",
		},
		{
			name:     "обратная числовая сортировка",
			input:    "cherry\t8\tFeb\nbanana\t3\tMar\napple\t5\tJan\n",
			args:     []string{"-k", "2", "-n", "-r"},
			expected: "cherry\t8\tFeb\napple\t5\tJan\nbanana\t3\tMar\n",
		},
		{
			name:     "комбинированные флаги",
			input:    "apple\t5\tJan\nbanana\t3\tMar\napple\t5\tJan\ncherry\t8\tFeb\n",
			args:     []string{"-k", "2", "-n", "-u"},
			expected: "banana\t3\tMar\napple\t5\tJan\ncherry\t8\tFeb\n",
		},
		{
			name:     "отсортированные данные",
			input:    "apple\nbanana\ncherry\n",
			args:     []string{"-c"},
			expected: "Данные отсортированы\n",
		},
		{
			name:     "неотсортированные данные",
			input:    "cherry\napple\nbanana\n",
			args:     []string{"-c"},
			expected: "Данные не отсортированы\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := runSort(test.input, test.args)
			if result != test.expected {
				t.Errorf("ожидалось %q, получено %q", test.expected, result)
			}
		})
	}
}

// runSort запускает программу sort с заданными аргументами
func runSort(input string, args []string) string {
	// Создаём временный файл с входными данными
	tmpFile, err := os.CreateTemp("", "sort_test_")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	// Записываем данные в файл
	_, err = tmpFile.WriteString(input)
	if err != nil {
		panic(err)
	}
	tmpFile.Close()

	// Запускаем программу
	cmdArgs := append(args, tmpFile.Name())
	cmd := exec.Command("./task10.exe", cmdArgs...)

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return string(output)
}

// TestMain проверяет, что программа собирается
func TestMain(m *testing.M) {
	// Проверяем, что программа собирается
	cmd := exec.Command("go", "build", "task10.go")
	if err := cmd.Run(); err != nil {
		panic("Программа не собирается: " + err.Error())
	}

	os.Exit(m.Run())
}
