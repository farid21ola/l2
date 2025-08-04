package main

import (
	"fmt"
	"sync"
	"time"
)

// or - объединяет done-каналы в один, при закрытии одного из каналов,
// закрывается и выходной канал
func or(channels ...<-chan interface{}) <-chan interface{} {
	out := make(chan interface{})
	var once sync.Once // Гарантирует, что close будет вызван только один раз

	for _, channel := range channels {
		go func(c <-chan interface{}) {
			<-c
			once.Do(func() {
				close(out)
			})
		}(channel)
	}
	return out
}

func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(2*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v\n", time.Since(start)) // done after 1.0014318s
}
