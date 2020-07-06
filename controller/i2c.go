package controller

import (
	"github.com/get-code-ch/mcp23008"
	"goswitch/model"
	"strconv"
)

func (device *Device) SetGPIO(data interface{}) {

	arguments := data.(map[string]interface{})

	module := arguments["module"].(string)
	sw, err := strconv.Atoi(arguments["sw"].(string))
	if err != nil {
		sw = -1
	}

	state, err := strconv.Atoi(arguments["state"].(string))
	if err != nil {
		state = -1
	}

	for idx := range device.Modules {
		if device.Modules[idx].Name == module && device.I2cMode == model.REAL && sw > -1 && state > -1 {
			if state == 0 {
				mcp23008.GpioOff(&device.Modules[idx], byte(sw))
			} else {
				mcp23008.GpioOn(&device.Modules[idx], byte(sw))
			}
			break
		}
	}

}

func (device *Device) GetGPIO(data interface{}) {
}
