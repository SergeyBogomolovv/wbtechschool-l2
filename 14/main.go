package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("done after %v", time.Since(start))
}

func or(channels ...<-chan any) <-chan any {
	c := make(chan any)

	go func() {
		defer close(c)
		switch len(channels) {
		case 0:
		case 1:
			<-channels[0]
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		case 3:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			}
		default:
			mid := len(channels) / 2
			select {
			case <-or(channels[:mid]...):
			case <-or(channels[mid:]...):
			}
		}
	}()

	return c
}

func sig(after time.Duration) <-chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}
