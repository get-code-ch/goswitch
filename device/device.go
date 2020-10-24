package main

import (
	"log"
	"os"
	"time"
)

func main() {
	// Wait group

	// Reading command line argument
	configFile := ""
	if len(os.Args) == 2 {
		configFile = os.Args[1]
	}
	device := NewDevice(configFile)

	log.Printf("Device %s initialized...\n", device.config.Name)
	channel := make(chan int)

	// Listening message from CommCenter
	go device.commService.Listen(channel, device)

	// Log a message every 15 minutes
	go func() {
		count := 1
		for {
			time.Sleep(15 * time.Minute)
			log.Printf("%d - Timer raised", count)
			count++
		}
	}()

	// Waiting end of go routine
	_ = <-channel

}
