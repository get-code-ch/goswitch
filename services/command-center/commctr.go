package main

import (
	"goswitch/config"
	"goswitch/controller"
	"log"
)


func main() {
	c := config.NewCommCtrConfig("")
	log.Printf("Config loaded... %v", c)
	controller.NewCommandCenter(c)

}
