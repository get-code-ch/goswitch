package main

import (
	"goswitch/config"
	"goswitch/controller"
	"log"
)

func main() {

	receiver := make(chan int)

	conf := config.NewDeviceConfig("")
	log.Printf("Config loaded... %v", conf.Controller)
	c := controller.NewDevice(conf)
	go c.Listen(receiver)
	c.Send("Message")
	<- receiver
}
