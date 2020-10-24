package main

import (
	"github.com/get-code-ch/ads1115"
	"github.com/get-code-ch/goswitch/common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"os"
	"strconv"
)

type ICService struct {
	Address     int
	Name        string
	Description string
	Type        common.IcRef
	IC          interface{}
	interrupt   chan byte
	commService *CommService
	endPoints   []common.Endpoint
}

// Init communication and IC
func (ic *ICService) InitIC(config common.IC, i2cDevice string, commService *CommService) {
	ic.Address = config.Address
	ic.Type = config.Type
	ic.interrupt = make(chan byte)
	ic.commService = commService
	ic.Description = config.Description
	ic.Name = config.Name

	switch ic.Type {
	// Initialize MCP23008 IO expander
	case common.MCP23008:

		// Initializing i2C bus
		if mcp, err := mcp23008.New(i2cDevice, ic.Name, ic.Address, 0, ic.Description); err == nil {
			ic.IC = &mcp
			go mcp23008.RegisterInterrupt(ic.IC.(*mcp23008.Mcp23008), ic.interrupt)
			go ic.listenMcp23008Interrupt(ic.interrupt)

			// Setting operation direction for GPIO
			for _, endPoint := range config.Endpoints {
				ic.endPoints = append(ic.endPoints, endPoint)
				if endPoint.Attributes["mode"] == "push" || endPoint.Attributes["mode"] == "input" {
					if gpio, err := strconv.Atoi(endPoint.Id); err == nil {
						mcp23008.GpioSetRead(ic.IC.(*mcp23008.Mcp23008), byte(gpio))
					} else {
						log.Printf("Invalid GPIO address format for endpoint %s\n", endPoint.Name)
					}
				}
			}

		} else {
			log.Printf("Error initializing IC %d -> %v", ic.Address, err)
		}

		break

	// Initialize Analog -> Digital converter ADS1115
	case common.ADS1115:
		ic.IC = new(*ads1115.Ads1115)
		for _, endPoint := range config.Endpoints {
			ic.endPoints = append(ic.endPoints, endPoint)
		}
		break

	}

}

func (ic *ICService) listenMcp23008Interrupt(interrupt chan byte) {
	for {
		gpio := int(<-interrupt)
		state := ic.readGPIO(gpio)
		log.Printf("Interrupt occurs on %d new state %d\n", gpio, state)
		endPoint := make(map[string]interface{})
		endPoint["address"] = ic.Address
		endPoint["gpio"] = gpio
		endPoint["id"] = gpio
		endPoint["state"] = state
		endPoint["value"] = state
		endPoint["type"] = common.MCP23008

		//endPoints := []map[string]interface{}{endPoint}

		msg := common.Message{Action: common.GPIOSTATE, Data: []map[string]interface{}{endPoint}, Client: common.Node{}, Server: ic.commService.me}

		ic.commService.Send(common.BROADCAST, msg)
	}
}

func (ic *ICService) readGPIO(gpio int) int {
	return int(mcp23008.ReadGpio(ic.IC.(*mcp23008.Mcp23008), byte(gpio)))
}

func (ic *ICService) GetInfo(device *Device, client common.Node) {

	action := common.BROADCAST
	if client.Id != "" && client.Type != "" {
		action = common.RELAY
	} else {
		client = common.Node{}
	}

	hostName, _ := os.Hostname()
	deviceInfo := make(map[string]interface{})
	deviceInfo["hostname"] = hostName
	deviceInfo["ic"] = ic.Name
	deviceInfo["address"] = ic.Address
	deviceInfo["endPoints"] = ic.endPoints
	deviceInfo["type"] = ic.Type
	deviceInfo["me"] = ic.commService.me
	info := common.Message{Action: common.SENDINFO, Data: deviceInfo, Client: client, Server: ic.commService.me}

	device.commService.Send(action, info)
}
