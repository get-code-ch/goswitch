package main

import (
	"github.com/get-code-ch/mcp23008"
	"goswitch/config"
	"goswitch/controller"
	"goswitch/model"
	"log"
	"os"
)

func main() {
	var err error

	receiver := make(chan int)

	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}

	conf := config.NewDeviceConfig(configFile)
	log.Printf("Config loaded... %v", conf.Controller)
	device := controller.NewDevice(conf)

	device.I2cMode = model.REAL
	device.Modules = make([]mcp23008.Mcp23008, len(conf.Modules))
	copy(device.Modules, conf.Modules)

	device.Switches = make([]config.I2cSwitch, len(conf.Switches))
	copy(device.Switches, conf.Switches)

	for idx := range device.Modules {

		device.Modules[idx], err = mcp23008.New(device.I2c, device.Modules[idx].Name, device.Modules[idx].Address, device.Modules[idx].Count, device.Modules[idx].Description)
		if err != nil {
			log.Printf("Error i2c module init -> %s", err)
			device.I2cMode = model.SIMULATION
		}
	}

	go device.Listen(receiver)
	<-receiver
}
