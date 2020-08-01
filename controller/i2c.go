package controller

import (
	"github.com/get-code-ch/mcp23008"
	"goswitch/model"
	"log"
	"math"
	"strconv"
	"strings"
)

func (device *Device) readGPIO(module *mcp23008.Mcp23008, gpio int) int {

	return int(mcp23008.ReadGpio(module, byte(gpio)))
}

func (device *Device) SetGPIO(data interface{}) {

	var err error

	request := data.(map[string]interface{})
	arguments := request["i2c"].(map[string]interface{})

	command := strings.ToLower(arguments["command"].(string))
	address, err := strconv.Atoi(arguments["address"].(string))
	if err != nil {

	}
	gpio, err := strconv.Atoi(arguments["gpio"].(string))
	if err != nil {
		gpio = -1
	}

	for idx := range device.Modules {
		if device.Modules[idx].Address == address && gpio > -1 {
			//			log.Printf("Module address-> %d / Address-> %d\n", device.Modules[idx].Address, address)
			//			log.Printf("Switch-> %d\n", byte(gpio))
			if device.I2cMode == model.REAL {
				switch strings.ToLower(command) {
				case "off":
					mcp23008.GpioOff(&device.Modules[idx], byte(gpio))
				case "on":
					mcp23008.GpioOn(&device.Modules[idx], byte(gpio))
				case "reverse":
					mcp23008.GpioReverse(&device.Modules[idx], byte(gpio))
				}
				state := device.readGPIO(&device.Modules[idx], gpio)
				for _, swc := range device.Switches {
					if swc.Address == device.Modules[idx].Address && swc.Gpio == gpio {
						swc.State = state
						swc.MacAddr = device.MacAddr
						msg := model.Message{Action: model.GPIOSTATE, Data: swc, Client: model.Node{}.SetFromInterface(request["client"]), Server: device.me}
						SendMessage(device, nil, model.BROADCAST, msg)
						break
					}
				}
			} else {
				action := 0
				switch command {
				case "off":
					log.Printf("Module %s switch %d switched Off\n", device.Modules[idx].Name, gpio)
					action = 0
				case "on":
					log.Printf("Module %s switch %d switched On\n", device.Modules[idx].Name, gpio)
					action = 1
				case "reverse":
					log.Printf("Module %s switch %d Reversed\n", device.Modules[idx].Name, gpio)
					action = -1
				}
				for _, swc := range device.Switches {
					if swc.Address == device.Modules[idx].Address && swc.Gpio == gpio {
						if action == -1 {
							swc.State = int(math.Abs(float64(swc.State + action)))
						} else {
							swc.State = action
						}
						swc.MacAddr = device.MacAddr
						msg := model.Message{Action: model.GPIOSTATE, Data: swc, Client: model.Node{}.SetFromInterface(request["client"]), Server: device.me}
						SendMessage(device, nil, model.BROADCAST, msg)
						break
					}
				}
			}
		}
	}
}

func (device *Device) GetAllGPIOState(data interface{}) {

	for _, swc := range device.Switches {
		for _, module := range device.Modules {
			if module.Address == swc.Address {
				if device.I2cMode == model.REAL {
					swc.State = int(mcp23008.ReadGpio(&module, byte(swc.Gpio)))
					swc.MacAddr = device.MacAddr
					msg := model.Message{Action: model.GPIOSTATE, Data: swc, Client: model.Node{Id: "", Type: model.CLI}, Server: device.me}
					SendMessage(device, nil, model.BROADCAST, msg)
				} else {
					log.Printf("Module: %s, Address: %d, GPIO_%d: %d", module.Name, swc.Address, swc.Gpio, swc.State)
				}
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
			if device.Modules[idx].MacAddr == module {
				if device.I2cMode == model.REAL {
					for gidx, gpio := range device.Modules[idx].Gpios {
						gpio = mcp23008.ReadGpio(&device.Modules[idx], byte(gidx))

					}
					info := model.Message{Action: model.SENDINFO, Data: state, Client: client}
					SendMessage(device, nil, model.RELAY, info)
				} else {
					log.Printf("Module %s switch %d Reversed\n", device.Modules[idx].MacAddr, sw)
				}
			}
		}
	*/
}
