package main

import (
	"fmt"
	"goswitch/config"
	"goswitch/controller"
	"log"
	"os"
)

func main() {

	msg := ""
	fmt.Printf("Hello %s\n", os.Args[1])

	receiver := make(chan int)

	conf := config.NewCliConfig("")
	log.Printf("Config loaded... %v", conf.Controller)
	c := controller.NewCli(conf)
	go controller.WaitMessages(c, receiver)

	for {
		fmt.Printf("> ")
		fmt.Scanf("%s", &msg)
		if msg == "..." {
			close(receiver)
			break
		}
		c.Echo(msg)
	}

	<-receiver

}
