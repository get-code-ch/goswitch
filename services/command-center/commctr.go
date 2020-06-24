package main

import (
	"goswitch/config"
	"goswitch/controller"
	"log"
)

func main() {
	receiver := make(chan int)

	c := config.NewCommCtrConfig("")
	log.Printf("Config loaded... %v", c)
	comCtr := controller.NewCommandCenter()
	go comCtr.Listen(c, receiver)
	<-receiver
}
