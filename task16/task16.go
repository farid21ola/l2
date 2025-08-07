package main

import (
	"fmt"
	"l2/task16/service"
)

func main() {
	url := "https://habr.com/ru/companies/vk/articles/314804/"
	depth := uint(1)

	folder := "pages"
	service := service.NewService(folder, depth)

	fmt.Printf("Начинаем рекурсивное сохранение сайта: %s\n", url)
	fmt.Printf("Максимальная глубина: %d\n", depth)

	err := service.Start(url)
	if err != nil {
		fmt.Println("Ошибка при сохранении:", err)
		return
	}

	fmt.Println("Сохранение завершено успешно!")
}
