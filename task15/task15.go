package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var (
	// curOS содержит название текущей операционной системы
	curOS = runtime.GOOS

	// unixToWindows содержит аналоги Unix-команд на Windows
	unixToWindows = map[string]string{
		"ps":   "tasklist",
		"grep": "findstr",
		"ls":   "dir",
	}

	// windowsToUnix содержит аналоги Windows-команд на Unix
	windowsToUnix = map[string]string{
		"tasklist": "ps",
		"findstr":  "grep",
		"dir":      "ls",
	}
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	runShell()
}

// runShell запускает основной цикл командного интерпретатора.
func runShell() {
	reader := bufio.NewReader(os.Stdin)

	for {
		currentDir, _ := os.Getwd()
		fmt.Printf("shell:%s$ ", currentDir)

		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF { // ctrl+d не работает в windows
				os.Exit(0)
			}
			fmt.Printf("Ошибка чтения ввода: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		executeCommand(input)
	}
}

// executeCommand обрабатывает и выполняет пользовательскую команду.
func executeCommand(input string) {
	parts := parseCommand(input)
	if len(parts) == 0 {
		return
	}

	if strings.Contains(input, "|") {
		executePipeline(input)
		return
	}
	executeAllCommand(parts)
}

// parseCommand разбирает строку команды на отдельные аргументы.
func parseCommand(input string) []string {
	return strings.Fields(input)
}

// executeAllCommand выполняет команду, определяя встроенная она или внешняя.
func executeAllCommand(parts []string) {
	switch parts[0] {
	case "cd":
		builtinCD(parts)
	case "pwd":
		builtinPWD()
	case "echo":
		builtinEcho(parts)
	case "kill":
		builtinKill(parts)
	case "ps":
		builtinPS()
	default:
		executeExternalCommand(parts)
	}
}

// builtinCD реализует встроенную команду смены текущей директории.
func builtinCD(parts []string) {
	if len(parts) < 2 {
		fmt.Println("Использование: cd <путь>")
		return
	}

	err := os.Chdir(parts[1])
	if err != nil {
		fmt.Printf("Ошибка смены директории: %v\n", err)
	}
}

// builtinPWD реализует встроенную команду вывода текущей рабочей директории.
func builtinPWD() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Ошибка получения текущей директории: %v\n", err)
		return
	}
	fmt.Println(currentDir)
}

// builtinEcho реализует встроенную команду вывода текста.
func builtinEcho(parts []string) {
	if len(parts) < 2 {
		fmt.Println()
		return
	}
	fmt.Println(strings.Join(parts[1:], " "))
}

// builtinKill реализует встроенную команду завершения процесса по PID.
func builtinKill(parts []string) {
	if len(parts) < 2 {
		fmt.Println("Использование: kill <pid>")
		return
	}

	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Printf("Неверный PID: %s\n", parts[1])
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Ошибка поиска процесса: %v\n", err)
		return
	}

	err = process.Kill()
	if err != nil {
		fmt.Printf("Ошибка завершения процесса: %v\n", err)
		return
	}

	fmt.Printf("Процесс %d завершен\n", pid)
}

// builtinPS реализует встроенную команду просмотра списка процессов.
func builtinPS() {
	task := "ps"
	if curOS == "windows" {
		v, ok := unixToWindows["ps"]
		if ok {
			task = v
		}
	}
	cmd := exec.Command(task)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка выполнения ps: %v\n", err)
	}
}

// executeExternalCommand выполняет внешнюю команду операционной системы.
func executeExternalCommand(parts []string) {
	if curOS == "windows" {
		v, ok := unixToWindows[parts[0]]
		if ok {
			parts[0] = v
		}
	} else {
		v, ok := windowsToUnix[parts[0]]
		if ok {
			parts[0] = v
		}
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Ошибка выполнения команды: %v\n", err)
	}
}

// executePipeline выполняет конвейер команд, соединенных символом "|".
func executePipeline(input string) {
	commands := strings.Split(input, "|")
	if len(commands) < 2 {
		fmt.Println("Ошибка: неверный синтаксис конвейера")
		return
	}

	var cmds []*exec.Cmd

	for _, commandStr := range commands {
		commandStr = strings.TrimSpace(commandStr)
		if commandStr == "" {
			continue
		}

		parts := parseCommand(commandStr)
		if len(parts) == 0 {
			continue
		}

		if curOS == "windows" {
			v, ok := unixToWindows[parts[0]]
			if ok {
				parts[0] = v
			}
		} else {
			v, ok := windowsToUnix[parts[0]]
			if ok {
				parts[0] = v
			}
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmds = append(cmds, cmd)
	}

	for i := 0; i < len(cmds)-1; i++ {
		stdout, err := cmds[i].StdoutPipe()
		if err != nil {
			fmt.Printf("ошибка создания пайпа: %v\n", err)
			return
		}
		cmds[i+1].Stdin = stdout
	}

	cmds[len(cmds)-1].Stdout = os.Stdout

	for _, cmd := range cmds {
		cmd.Stderr = os.Stderr
	}

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fmt.Printf("ошибка запуска команды: %v\n", err)
			return
		}
	}

	for _, cmd := range cmds {
		_ = cmd.Wait()
		// Игнорируем ошибки, так как некоторые команды могут возвращать
		// ненулевой код выхода при нормальной работе
	}
}
