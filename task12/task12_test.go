package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestGrepBasic проверяет базовый поиск текста
func TestGrepBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{
			name:     "базовый поиск",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\n",
			args:     []string{"hello"},
			expected: "Вторая строка с hello\n",
		},
		{
			name:     "поиск без учета регистра",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка с HELLO\n",
			args:     []string{"-i", "hello"},
			expected: "Вторая строка с hello\nЧетвертая строка с HELLO\n",
		},
		{
			name:     "подсчет строк",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка с HELLO\n",
			args:     []string{"-c", "hello"},
			expected: "1\n",
		},
		{
			name:     "номера строк",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\n",
			args:     []string{"-n", "hello"},
			expected: "2:Вторая строка с hello\n",
		},
		{
			name:     "инвертированный поиск",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка с HELLO\n",
			args:     []string{"-v", "-i", "hello"},
			expected: "Первая строка\nТретья строка\n",
		},
		{
			name:     "фиксированная строка",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка с HELLO\n",
			args:     []string{"-F", "hello"},
			expected: "Вторая строка с hello\n",
		},
		{
			name:     "контекст после",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка\n",
			args:     []string{"-A", "1", "hello"},
			expected: "Вторая строка с hello\n-Третья строка\n",
		},
		{
			name:     "контекст до",
			input:    "Первая строка\nВторая строка\nТретья строка с hello\nЧетвертая строка\n",
			args:     []string{"-B", "1", "hello"},
			expected: "-Вторая строка\nТретья строка с hello\n",
		},
		{
			name:     "контекст вокруг",
			input:    "Первая строка\nВторая строка\nТретья строка с hello\nЧетвертая строка\nПятая строка\n",
			args:     []string{"-C", "1", "hello"},
			expected: "-Вторая строка\nТретья строка с hello\n-Четвертая строка\n",
		},
		{
			name:     "комбинированные флаги",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\nЧетвертая строка с HELLO\n",
			args:     []string{"-i", "-n", "-c", "hello"},
			expected: "2\n",
		},
		{
			name:     "регулярное выражение",
			input:    "Первая строка\nВторая строка с hello\nТретья строка с world\n",
			args:     []string{"h.*o"},
			expected: "Вторая строка с hello\n",
		},
		{
			name:     "поиск из STDIN",
			input:    "Строка с hello\nДругая строка\n",
			args:     []string{"hello"},
			expected: "Строка с hello\n",
		},
		{
			name:     "поиск с номерами и контекстом",
			input:    "Первая строка\nВторая строка\nТретья строка с hello\nЧетвертая строка\n",
			args:     []string{"-n", "-A", "1", "hello"},
			expected: "3:Третья строка с hello\n4:-Четвертая строка\n",
		},
		{
			name:     "инвертированный поиск с номерами",
			input:    "Первая строка\nВторая строка с hello\nТретья строка\n",
			args:     []string{"-v", "-n", "hello"},
			expected: "1:Первая строка\n3:Третья строка\n",
		},
		{
			name:     "перекрывающиеся совпадения",
			input:    "Первая hello\nВторая hello\nТретья hello\n",
			args:     []string{"-C", "1", "hello"},
			expected: "Первая hello\nВторая hello\nТретья hello\n",
		},
		{
			name:     "совпадение в первой и последней строке",
			input:    "Первая hello\nСредняя строка\nПредпоследняя строка\nПоследняя hello\n",
			args:     []string{"-C", "1", "hello"},
			expected: "Первая hello\n-Средняя строка\n-Предпоследняя строка\nПоследняя hello\n",
		},
		{
			name:     "нет совпадений",
			input:    "Первая строка\nВторая строка\nТретья строка\n",
			args:     []string{"hello"},
			expected: "",
		},
		{
			name:     "только контекст, без совпадений",
			input:    "Первая строка\nВторая строка\nТретья строка\n",
			args:     []string{"-C", "1", "hello"},
			expected: "",
		},
		{
			name:     "инвертированный поиск с контекстом",
			input:    "Первая hello\nВторая строка\nТретья hello\nЧетвертая строка\nПятая hello\n",
			args:     []string{"-v", "-C", "1", "hello"},
			expected: "-Первая hello\nВторая строка\n-Третья hello\nЧетвертая строка\n-Пятая hello\n",
		},
		{
			name:     "-F с похожей на регулярку строкой",
			input:    "a.b\naab\nabb\n",
			args:     []string{"-F", "a.b"},
			expected: "a.b\n",
		},
		{
			name:     "-c с контекстом",
			input:    "Первая hello\nВторая строка\nТретья hello\n",
			args:     []string{"-c", "-C", "1", "hello"},
			expected: "2\n",
		},
		{
			name:     "-n с контекстом",
			input:    "Первая строка\nВторая hello\nТретья строка\n",
			args:     []string{"-n", "-C", "1", "hello"},
			expected: "1:-Первая строка\n2:Вторая hello\n3:-Третья строка\n",
		},
		{
			name:     "-v с -c",
			input:    "Первая строка\nВторая hello\nТретья строка\n",
			args:     []string{"-v", "-c", "hello"},
			expected: "2\n",
		},
		{
			name:     "пустой файл",
			input:    "",
			args:     []string{"hello"},
			expected: "",
		},
		{
			name:     "STDIN с контекстом",
			input:    "Первая строка\nВторая hello\nТретья строка\n",
			args:     []string{"-C", "1", "hello"},
			expected: "-Первая строка\nВторая hello\n-Третья строка\n",
		},
		{
			name:     "многострочные совпадения (регулярка)",
			input:    "строка hello\nстрока world\nстрока123hello\n",
			args:     []string{"строка.*hello"},
			expected: "строка hello\nстрока123hello\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := runGrep(test.input, test.args)
			if result != test.expected {
				t.Errorf("ожидалось %q, получено %q", test.expected, result)
			}
		})
	}
}

// runGrep запускает программу grep с заданными аргументами
func runGrep(input string, args []string) string {
	// Создаём временный файл с входными данными
	tmpFile, err := os.CreateTemp("", "grep_test_")
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
	cmd := exec.Command("./task12.exe", cmdArgs...)

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	return string(output)
}
