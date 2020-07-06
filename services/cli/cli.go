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
	listRE := regexp.MustCompile(`(?mi)^(list)\s?[\r\n]+$`)
	getInfoRE := regexp.MustCompile(`(?mi)^(info)\s([\S]+)\s?[\r\n]+$`)
	setGpioRE := regexp.MustCompile(`(?mi)^(set)\s(?P<id>[\S]+)\s(?P<module>[\S]+)\s(?P<sw>[\S]+)\s(?P<state>[0|1])\s?[\r\n]+$`)
	fakeRE := regexp.MustCompile(`^(fake)[\r\n]+$`)
	//getXxxRE := regexp.MustCompile(`(?mi)^(xxx)\s([\S]+)\s?[\r\n]+$`)

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
			//msg := listRE.FindStringSubmatch(input)[2]
			controller.SendMessage(cli, nil, model.LIST, "")
			continue
		}

		if getInfoRE.MatchString(input) {
			device := getInfoRE.FindStringSubmatch(input)[2]
			data := model.Message{Action: model.GETINFO, Client: model.Node{Type: model.DEVICE, Id: device}, Data: model.Node{Type: model.CLI, Id: cli.Name}}
			controller.SendMessage(cli, nil, model.RELAY, data)
			continue
		}

		if setGpioRE.MatchString(input) {
			command := setGpioRE.FindStringSubmatch(input)
			arguments := make(map[string]interface{})

			for idx, name := range setGpioRE.SubexpNames() {
				if idx != 0 && name != "" {
					arguments[name] = command[idx]
				}
			}
			data := model.Message{Action: model.SETGPIO, Client: model.Node{Type: model.DEVICE, Id: arguments["id"].(string)}, Data: arguments}
			controller.SendMessage(cli, nil, model.RELAY, data)
			continue
		}

		if fakeRE.MatchString(input) {
			controller.SendMessage(cli, nil, model.FAKE, "--Fake datas--")
			continue
		}

		/*
			if getXxxRE.MatchString(input) {
				device := getXxxRE.FindStringSubmatch(input)[2]
				data := model.Message{Action: model.xxx, Client: model.Node{Type: model.DEVICE, Id: device}, Data: model.Node{Type: model.CLI, Id: cli.Name}}
				controller.SendMessage(cli, nil, model.RELAY, data)
				continue
			}
		*/

		msg := input[:len(input)-1]
		controller.SendMessage(cli, nil, model.ECHO, msg)
	}

	<-receiver

}
