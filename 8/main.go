package main

import (
	"fmt"
	"log"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	time, err := CurrentTime("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Current time is: %v", time)
}

// CurrentTime returns current time from ntp server.
func CurrentTime(address string) (time.Time, error) {
	currTime, err := ntp.Time(address)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get time: %w", err)
	}
	return currTime, nil
}
