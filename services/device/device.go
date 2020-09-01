package main

import (
	"github.com/get-code-ch/ads1115"
	"github.com/get-code-ch/goswitch/config"
	"github.com/get-code-ch/goswitch/controller"
	"github.com/get-code-ch/goswitch/model"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"os"
	"time"
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

	device.Modules = make(map[int]mcp23008.Mcp23008)
	for key, value := range conf.Modules {
		device.Modules[key] = value
	}

	device.Switches = make([]config.I2cSwitch, len(conf.Switches))
	copy(device.Switches, conf.Switches)

	for idx := range device.Switches {
		device.Switches[idx].MacAddr = device.MacAddr
	}

	for idx, module := range device.Modules {

		device.Modules[idx], err = mcp23008.New(device.I2c, module.Name, module.Address, module.Count, module.Description)
		if err != nil {
			log.Printf("Error i2c module init -> %s", err)
			device.I2cMode = model.SIMULATION
		}
	}

	go func() {
		if ads, err := ads1115.New(device.I2c, "", 0x48, ""); err != nil {
			log.Printf("Error initializing ADS1115 --> %v\n", err)
		} else {
			for {
				v, i, b := ads1115.ReadConversionRegister(&ads)
				log.Printf("Conversion register value --> [%d] %.4f | %v\n", v, i, b)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	go device.Listen(receiver)
	<-receiver
}
