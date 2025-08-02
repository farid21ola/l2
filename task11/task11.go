package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	arr := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(SearchAnagramms(arr))
	// map[листок:[листок слиток столик] пятак:[пятак пятка тяпка]]
	arr1 := []string{"пятка", "пятак", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(SearchAnagramms(arr1))
	// map[листок:[листок слиток столик] пятка:[пятак пятка тяпка]]
}

// SearchAnagramms находит группы анаграмм в массиве строк.
func SearchAnagramms(arr []string) map[string][]string {
	res := make(map[string][]string)
	sortedMap := make(map[string]string) // [sorted]original

	// Находим анаграммы путём сортировки букв в словах
	for i, x := range arr {
		if i == 0 {
			x = strings.ToLower(x)
			res[x] = make([]string, 0, 1)
			res[x] = append(res[x], x)

			sorted := sortStrings(x)
			sortedMap[sorted] = x
		} else {
			x = strings.ToLower(x)

			sorted := sortStrings(x)

			orig, ok := sortedMap[sorted]
			if ok {
				res[orig] = append(res[orig], x)
			} else {
				res[x] = make([]string, 0, 1)
				res[x] = append(res[x], x)
				sortedMap[sorted] = x
			}
		}
	}

	// Удаляем множества из 1 слова и сортируем остальные
	for k, v := range res {
		if len(v) == 1 {
			delete(res, k)
		} else {
			sort.Strings(v)
		}
	}
	return res
}

// sortStrings сортирует символы в строке по возрастанию.
func sortStrings(str string) string {
	runes := []rune(str)

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})

	return string(runes)
}
