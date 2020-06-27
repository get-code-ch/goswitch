package main

import (
	"bufio"
	"fmt"
	"goswitch/config"
	"goswitch/controller"
	"goswitch/model"
	"log"
	"os"
	"regexp"
)

func main() {

	endRE := regexp.MustCompile(`^\.{3}[\r\n]+$`)
	listRE := regexp.MustCompile(`(?mi)^(List)(?:\s(.*))?[\r\n]+$`)
	getInfoRE := regexp.MustCompile(`(?mi)^(Info)\s(.*)?[\r\n]+$`)

	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}

	receiver := make(chan int)

	conf := config.NewCliConfig(configFile)
	log.Printf("Config loaded... %v\n", conf.Controller)
	fmt.Printf("\n> ")

	cli := controller.NewCli(conf)
	go controller.WaitMessages(cli, receiver)

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		if endRE.MatchString(input) {
			close(receiver)
			break
		}

		if listRE.MatchString(input) {
			msg := listRE.FindStringSubmatch(input)[2]
			controller.SendMessage(cli, nil, model.LIST, msg)
			continue
		}

		if getInfoRE.MatchString(input) {
			device := getInfoRE.FindStringSubmatch(input)[2]
			data := model.Message{Action: model.GETINFO, Client: model.Node{Type: model.DEVICE, Id: device}, Data: model.Node{Type: model.CLI, Id: cli.Name}}
			controller.SendMessage(cli, nil, model.RELAY, data)

		}

		msg := input[:len(input)-1]
		controller.SendMessage(cli, nil, model.ECHO, msg)
	}

	<-receiver

}
