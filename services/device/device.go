package main

import (
	"goswitch/config"
	"goswitch/controller"
	"log"
	"os"
)

func main() {
	receiver := make(chan int)

	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}

	conf := config.NewDeviceConfig(configFile)
	log.Printf("Config loaded... %v", conf.Controller)
	device := controller.NewDevice(conf)
	go device.Listen(receiver)
	<-receiver
}
