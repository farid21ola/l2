package main

import (
	"fmt"
	"log"
	"time"

	"github.com/beevik/ntp"
)

// Можно добавить больше адресов для лучшей отказоустойчивости
var ipAddresses = [...]string{
	"0.beevik-ntp.pool.ntp.org",
	"1.beevik-ntp.pool.ntp.org",
	"2.beevik-ntp.pool.ntp.org",
	"3.beevik-ntp.pool.ntp.org",
}

func main() {
	if err := PrintCurrentTime(); err != nil {
		log.Fatal(err)
	}
}

// PrintCurrentTime получает текущее время с одного из NTP-серверов и выводит в консоль.
func PrintCurrentTime() error {
	currentTime, err := CurrentTime()
	if err != nil {
		return err
	}
	fmt.Printf("Current time: %s\n", currentTime)
	return nil
}

// CurrentTime получает текущее время с одного из NTP-серверов.
func CurrentTime() (time.Time, error) {
	var lastErr error
	for _, url := range ipAddresses {
		currentTime, err := ntp.Time(url)
		if err == nil {
			return currentTime, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("failed to get current time from all ntp-servers: %w", lastErr)
}
