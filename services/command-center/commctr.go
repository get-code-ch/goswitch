package main

import (
	"github.com/get-code-ch/goswitch/config"
	"github.com/get-code-ch/goswitch/controller"
	"log"
	"os"
)

func main() {
	receiver := make(chan int)

	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}

	c := config.NewCommCtrConfig(configFile)
	log.Printf("Config loaded... %v", c)
	comCtr := controller.NewCommandCenter(c)
	go controller.WaitMessages(comCtr, receiver)
	<-receiver
}
