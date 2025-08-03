package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// GrepOptions содержит все опции для поиска текста
type GrepOptions struct {
	afterContext  int    // -A N
	beforeContext int    // -B N
	context       int    // -C N
	countOnly     bool   // -c
	ignoreCase    bool   // -i
	invertMatch   bool   // -v
	fixedString   bool   // -F
	lineNumber    bool   // -n
	pattern       string // шаблон поиска
}

// Line представляет строку текста с номером
type Line struct {
	number int    // номер строки
	text   string // содержимое строки
}

// MatchResult содержит результат поиска для одной строки
type MatchResult struct {
	line    Line   // найденная строка
	isMatch bool   // флаг совпадения
	context []Line // контекст вокруг строки
}

func main() {
	// Определение флагов
	afterContext := flag.Int("A", 0, "вывести N строк после каждой найденной строки")
	beforeContext := flag.Int("B", 0, "вывести N строк до каждой найденной строки")
	context := flag.Int("C", 0, "вывести N строк контекста вокруг найденной строки")
	countOnly := flag.Bool("c", false, "выводить только количество совпадающих строк")
	ignoreCase := flag.Bool("i", false, "игнорировать регистр")
	invertMatch := flag.Bool("v", false, "инвертировать фильтр")
	fixedString := flag.Bool("F", false, "воспринимать шаблон как фиксированную строку")
	lineNumber := flag.Bool("n", false, "выводить номер строки")

	flag.Parse()

	// Получение шаблона из аргументов
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: не указан шаблон для поиска")
		flag.Usage()
		os.Exit(1)
	}

	pattern := args[0]
	var inputFile string
	if len(args) > 1 {
		inputFile = args[1]
	}

	// Создание опций
	opts := &GrepOptions{
		afterContext:  *afterContext,
		beforeContext: *beforeContext,
		context:       *context,
		countOnly:     *countOnly,
		ignoreCase:    *ignoreCase,
		invertMatch:   *invertMatch,
		fixedString:   *fixedString,
		lineNumber:    *lineNumber,
		pattern:       pattern,
	}

	// Применение контекста
	if opts.context > 0 {
		opts.afterContext = opts.context
		opts.beforeContext = opts.context
	}

	// Определение источника ввода
	var reader io.Reader
	if inputFile == "" {
		reader = os.Stdin
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка открытия файла: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	}

	// Выполнение поиска
	matches, err := grep(reader, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при поиске: %v\n", err)
		os.Exit(1)
	}

	// Вывод результатов
	if opts.countOnly {
		fmt.Println(len(matches))
	} else {
		printResults(matches, opts)
	}
}

// grep выполняет поиск текста в потоке данных
func grep(reader io.Reader, opts *GrepOptions) ([]MatchResult, error) {
	var matches []MatchResult
	var allLines []Line
	var matchLines []int

	// Чтение всех строк
	scanner := bufio.NewScanner(reader)
	lineNum := 1
	for scanner.Scan() {
		line := Line{
			number: lineNum,
			text:   scanner.Text(),
		}
		allLines = append(allLines, line)
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Компиляция регулярного выражения или создание строки для поиска
	var searchPattern string
	var regex *regexp.Regexp
	var err error

	if opts.fixedString {
		searchPattern = opts.pattern
		if opts.ignoreCase {
			searchPattern = strings.ToLower(searchPattern)
		}
	} else {
		pattern := opts.pattern
		if opts.ignoreCase {
			pattern = "(?i)" + pattern
		}
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("неверное регулярное выражение: %v", err)
		}
	}

	// Поиск совпадений
	for i, line := range allLines {
		var isMatch bool

		if opts.fixedString {
			lineText := line.text
			if opts.ignoreCase {
				lineText = strings.ToLower(lineText)
			}
			isMatch = strings.Contains(lineText, searchPattern)
		} else {
			isMatch = regex.MatchString(line.text)
		}

		// Инвертирование результата если указан флаг -v
		if opts.invertMatch {
			isMatch = !isMatch
		}

		if isMatch {
			matchLines = append(matchLines, i)
		}
	}

	// Создание результатов с контекстом
	for _, matchIndex := range matchLines {
		result := MatchResult{
			line:    allLines[matchIndex],
			isMatch: true,
		}

		// Добавление контекста
		if opts.afterContext > 0 || opts.beforeContext > 0 {
			result.context = getContext(allLines, matchIndex, opts.beforeContext, opts.afterContext)
		}

		matches = append(matches, result)
	}

	return matches, nil
}

// getContext возвращает контекст вокруг найденной строки
func getContext(allLines []Line, matchIndex, before, after int) []Line {
	var context []Line

	// Добавление строк до
	start := matchIndex - before
	if start < 0 {
		start = 0
	}

	// Добавление строк после
	end := matchIndex + after + 1
	if end > len(allLines) {
		end = len(allLines)
	}

	// Собираем контекст
	for i := start; i < end; i++ {
		if i != matchIndex { // Исключаем саму найденную строку
			context = append(context, allLines[i])
		}
	}

	return context
}

// printResults выводит результаты поиска в стандартный вывод
func printResults(matches []MatchResult, opts *GrepOptions) {
	lastPrintedLine := -1

	// Собираем номера всех совпадающих строк
	matchNumbers := make(map[int]bool)
	for _, match := range matches {
		matchNumbers[match.line.number] = true
	}

	for _, match := range matches {
		if match.line.number <= lastPrintedLine {
			fmt.Println("|||", match.line.number, lastPrintedLine)
			continue
		}

		var beforeContext, afterContext []Line
		if len(match.context) > 0 {
			matchLineNum := match.line.number
			for _, contextLine := range match.context {
				if contextLine.number < matchLineNum {
					beforeContext = append(beforeContext, contextLine)
				} else {
					afterContext = append(afterContext, contextLine)
				}
			}
		}

		// Вывод контекста до
		for _, contextLine := range beforeContext {
			if contextLine.number > lastPrintedLine {
				printLine(contextLine, opts, false)
				lastPrintedLine = contextLine.number
			}
		}

		// Вывод найденной строки
		printLine(match.line, opts, true)
		lastPrintedLine = match.line.number

		// Вывод контекста после
		for _, contextLine := range afterContext {
			if contextLine.number > lastPrintedLine {
				if matchNumbers[contextLine.number] {
					break
				} else {
					printLine(contextLine, opts, false)
				}
				lastPrintedLine = contextLine.number
			}
		}
	}
}

// printLine выводит одну строку с учётом настроек форматирования
func printLine(line Line, opts *GrepOptions, isMatch bool) {
	prefix := ""

	if opts.lineNumber {
		prefix = fmt.Sprintf("%d:", line.number)
	}

	if isMatch {
		fmt.Printf("%s%s\n", prefix, line.text)
	} else {
		// Для контекста добавляем дефис
		fmt.Printf("%s-%s\n", prefix, line.text)
	}
}
