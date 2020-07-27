package controller

import (
	"github.com/get-code-ch/mcp23008"
	"goswitch/model"
	"log"
	"strconv"
	"strings"
)

func (device *Device) SetGPIO(data interface{}) {

	var err error
	arguments := data.(map[string]interface{})

	command := strings.ToLower(arguments["command"].(string))
	address, err := strconv.Atoi(arguments["address"].(string))
	if err != nil {

	}
	sw, err := strconv.Atoi(arguments["gpio"].(string))
	if err != nil {
		sw = -1
	}

	for idx := range device.Modules {
		if device.Modules[idx].Address == address && sw > -1 {
			log.Printf("Module address-> %d / Address-> %d\n", device.Modules[idx].Address, address)
			log.Printf("Switch-> %d\n", byte(sw))
			if device.I2cMode == model.REAL {
				switch strings.ToLower(command) {
				case "off":
					mcp23008.GpioOff(&device.Modules[idx], byte(sw))
				case "on":
					mcp23008.GpioOn(&device.Modules[idx], byte(sw))
				case "reverse":
					mcp23008.GpioReverse(&device.Modules[idx], byte(sw))
				}
				break
			} else {
				switch command {
				case "off":
					log.Printf("Module %s switch %d switched Off\n", device.Modules[idx].Name, sw)
				case "on":
					log.Printf("Module %s switch %d switched On\n", device.Modules[idx].Name, sw)
				case "reverse":
					log.Printf("Module %s switch %d Reversed\n", device.Modules[idx].Name, sw)
				}
				break
			}
		}
	}
}

func (device *Device) GetGPIO(data interface{}) {
	/*
		client := model.Node{}.SetFromInterface(data)

		arguments := data.(map[string]interface{})
		module := arguments["module"].(string)

		for idx := range device.Modules {
			if device.Modules[idx].Name == module {
				if device.I2cMode == model.REAL {
					for gidx, gpio := range device.Modules[idx].Gpios {
						gpio = mcp23008.ReadGpio(&device.Modules[idx], byte(gidx))

					}
					info := model.Message{Action: model.SENDINFO, Data: state, Client: client}
					SendMessage(device, nil, model.RELAY, info)
				} else {
					log.Printf("Module %s switch %d Reversed\n", device.Modules[idx].Name, sw)
				}
			}
		}
	*/
}
