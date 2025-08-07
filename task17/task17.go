package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Пример использования:
// go run task17.go --timeout 15s smtp.gmail.com 25

func main() {
	timeoutFlag := flag.Duration("timeout", 10*time.Second, "timeout for connection")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Usage: go run task17.go [--timeout duration] host port")
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	timeout := *timeoutFlag
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		fmt.Printf("Ошибка подключения к %s:%s: %v\n", host, port, err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Подключено к %s:%s\n", host, port)

	done := make(chan struct{}, 1)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("Получен сигнал, закрываем соединение\n")
		done <- struct{}{}
	}()

	// Горутина для копирования данных из сокета в STDOUT
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil && err != io.EOF {
			fmt.Printf("Ошибка чтения с сервера: %v\n", err)
			done <- struct{}{}
			return
		} else if err == io.EOF {
			fmt.Printf("Сервер закрыл соединение\n")
			done <- struct{}{}
			return
		}
	}()

	// Горутина для копирования данных из STDIN в сокет
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil && err != io.EOF {
			fmt.Printf("Ошибка записи на сервер: %v\n", err)
			done <- struct{}{}
			return
		}
	}()

	<-done
	fmt.Printf("Соединение закрыто\n")
}
