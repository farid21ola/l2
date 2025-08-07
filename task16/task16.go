package main

import (
	"flag"
	"fmt"
	"l2/task16/service"
	"os"
)

func main() {
	var depth uint
	var folder string

	flag.UintVar(&depth, "depth", 1, "Максимальная глубина рекурсии")
	flag.StringVar(&folder, "folder", "pages", "Папка для сохранения файлов")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Использование: task16 [флаги] <URL>")
		fmt.Println("Флаги:")
		flag.PrintDefaults()
		fmt.Println("Пример: task16 -depth=2 https://habr.com/ru/companies/vk/articles/314804/")
		os.Exit(1)
	}

	url := args[0]

	if url == "" {
		fmt.Println("Ошибка: URL не может быть пустым")
		os.Exit(1)
	}

	service := service.NewService(folder, depth)

	fmt.Printf("Начинаем рекурсивное сохранение сайта: %s\n", url)
	fmt.Printf("Максимальная глубина: %d\n", depth)
	fmt.Printf("Папка для сохранения: %s\n", folder)

	err := service.Start(url)
	if err != nil {
		fmt.Println("Ошибка при сохранении:", err)
		os.Exit(1)
	}

	fmt.Println("Сохранение завершено успешно!")
}
