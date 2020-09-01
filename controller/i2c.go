package controller

import (
	"github.com/get-code-ch/goswitch/model"
	"github.com/get-code-ch/mcp23008/v3"
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

	if module, ok := device.Modules[address]; ok {
		if device.I2cMode == model.REAL {
			switch strings.ToLower(command) {
			case "off":
				mcp23008.GpioOff(&module, byte(gpio))
			case "on":
				mcp23008.GpioOn(&module, byte(gpio))
			case "reverse":
				mcp23008.GpioReverse(&module, byte(gpio))
			}
			state := device.readGPIO(&module, gpio)
			for _, swc := range device.Switches {
				if swc.Address == module.Address && swc.Gpio == gpio {
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
				log.Printf("Module %s switch %d switched Off\n", module.Name, gpio)
				action = 0
			case "on":
				log.Printf("Module %s switch %d switched On\n", module.Name, gpio)
				action = 1
			case "reverse":
				log.Printf("Module %s switch %d Reversed\n", module.Name, gpio)
				action = -1
			}
			for iSwc, swc := range device.Switches {
				if swc.Address == module.Address && swc.Gpio == gpio {
					if action == -1 {
						swc.State = int(math.Abs(float64(swc.State + action)))
					} else {
						swc.State = action
					}
					device.Switches[iSwc].State = swc.State
					msg := model.Message{Action: model.GPIOSTATE, Data: swc, Client: model.Node{}.SetFromInterface(request["client"]), Server: device.me}
					SendMessage(device, nil, model.BROADCAST, msg)
					break
				}
			}
		}
	}
}

func (device *Device) GetAllGPIOState(data interface{}) {

	for _, swc := range device.Switches {
		if module, ok := device.Modules[swc.Address]; ok {
			if device.I2cMode == model.REAL {
				swc.State = int(mcp23008.ReadGpio(&module, byte(swc.Gpio)))
				swc.MacAddr = device.MacAddr
				msg := model.Message{Action: model.GPIOSTATE, Data: swc, Client: model.Node{Id: "", Type: model.CLI}, Server: device.me}
				SendMessage(device, nil, model.BROADCAST, msg)
			} else {
				//					log.Printf("Module: %s, Address: %d, GPIO_%d: %d", module.Name, swc.Address, swc.Gpio, swc.State)
				log.Printf("Module: %s, Address: %d, GPIO_%d: %d", module.Name, swc.Address, swc.Gpio, swc.State)
			}
		}
	}
}
