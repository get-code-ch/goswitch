package main

import (
	"log"
	"os"
)

func main() {
	receiver := make(chan int)

	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}

	c := NewCommCtrConfig(configFile)
	log.Printf("Config loaded... %v", c)
	comCtr := NewCommandCenter(c)
	go comCtr.Listen(receiver)
	<-receiver
}
