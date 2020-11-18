package main

import (
	"fmt"
	"github.com/get-code-ch/ads1115"
	"github.com/get-code-ch/goswitch/common"
	"github.com/get-code-ch/mcp23008/v3"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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
				endPoint.RefreshInterval = -1
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
		if ads, err := ads1115.New(i2cDevice, "", ic.Address, ""); err == nil {
			ic.IC = &ads
			for _, endPoint := range config.Endpoints {
				ic.endPoints = append(ic.endPoints, endPoint)

				if endPoint.RefreshInterval > 0 {
					go func(ep common.Endpoint) {
						for {
							time.Sleep(time.Duration(ep.RefreshInterval) * time.Second)
							ic.refreshADSValue(ep.Id)
						}
					}(endPoint)
				}
			}
		}
		break

	}

}

func (ic *ICService) listenMcp23008Interrupt(interrupt chan byte) {
	for {
		// Waiting PIN interruption and reading value
		gpio := int(<-interrupt)
		state := ic.readGPIO(gpio)

		// Broadcasting new state to all clients
		endpointMap := make(map[string]interface{})
		endpointMap["address"] = ic.Address
		endpointMap["gpio"] = gpio
		endpointMap["id"] = gpio
		endpointMap["state"] = state
		endpointMap["value"] = state
		endpointMap["type"] = common.MCP23008

		msg := common.Message{Action: common.GPIOSTATE, Data: []map[string]interface{}{endpointMap}, Client: common.Node{}, Server: ic.commService.me}
		ic.commService.Send(common.BROADCAST, msg)

		// Reading attributes of endpoint
		endpoint := common.Endpoint{Id: ""}
		for _, ep := range ic.endPoints {
			if ep.Id == strconv.Itoa(gpio) {
				endpoint = ep
				break
			}
		}

		// Check if notification must be sent to Telegram
		if endpoint.Telegram != (common.Telegram{}) && endpoint.Telegram.Notification == common.ONCHANGE {
			msg = common.Message{Action: common.TO_TME,
				Data:   fmt.Sprintf("State changed for %s - %s -> %d", ic.Name, endpoint.Name, state),
				Client: common.Node{},
				Server: ic.commService.me}
			ic.commService.Send(common.TO_TME, msg)
		}

		// Checking if endpoint had a slave and updating state of slave
		if slave, ok := endpoint.Attributes["slave"]; ok {
			if gpio, err := strconv.Atoi(slave.(string)); err == nil {
				if mode, ok := endpoint.Attributes["mode"]; ok {
					log.Printf("mode --> %s for %d", mode, gpio)
					if strings.ToLower(mode.(string)) == "push" {
						state = int(math.Abs(float64(ic.readGPIO(gpio) - 1)))
						log.Printf("push state %d for %d", state, gpio)
					}
				}
				ic.writeGPIO(gpio, state)
				go func() {
					log.Printf("New state for gpio %d is %d", gpio, state)
					ic.interrupt <- byte(gpio)
				}()
			}
		}
	}
}

func (ic *ICService) readGPIO(gpio int) int {
	return int(mcp23008.ReadGpio(ic.IC.(*mcp23008.Mcp23008), byte(gpio)))
}

func (ic *ICService) writeGPIO(gpio int, state int) int {
	if state == 0 {
		mcp23008.GpioOff(ic.IC.(*mcp23008.Mcp23008), byte(gpio))
	} else {
		mcp23008.GpioOn(ic.IC.(*mcp23008.Mcp23008), byte(gpio))
	}
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

func (ic *ICService) refreshADSValue(id string) {
	value := ic.readValue(id)
	currentEp := common.Endpoint{Id: ""}
	//log.Printf("VIn is %f for %s\n", value, id)

	endPoint := make(map[string]interface{})
	endPoint["address"] = ic.Address
	endPoint["id"] = id
	endPoint["type"] = common.ADS1115
	for _, ep := range ic.endPoints {
		if ep.Id == id {
			endPoint["attributes"] = ep.Attributes
			currentEp = ep
			break
		}
	}

	endPoint["value"] = value

	if currentEp.Id == id {
		msg := common.Message{Action: common.DIGITALVALUE, Data: []map[string]interface{}{endPoint}, Client: common.Node{}, Server: ic.commService.me}
		ic.commService.Send(common.BROADCAST, msg)

		if currentEp.Telegram != (common.Telegram{}) && currentEp.Telegram.Notification != common.NO {
			msg = common.Message{Action: common.TO_TME,
				Data:   fmt.Sprintf("Value for %s - %s -> %.2f %s", ic.Name, currentEp.Name, value, currentEp.Attributes["unit"]),
				Client: common.Node{},
				Server: ic.commService.me}
			ic.commService.Send(common.TO_TME, msg)
		}
	}
}

func (ic *ICService) readValue(endpoint string) float64 {
	ads := ic.IC.(*ads1115.Ads1115)
	vIn := ads1115.ReadConversionRegister(ads, endpoint)
	result := 0.0

	for _, ep := range ic.endPoints {
		if ep.Id == endpoint {
			if _, ok := ep.Attributes["scale"]; ok {
				if _, ok := ep.Attributes["convert"]; ok {
					fnc := reflect.ValueOf(ic).MethodByName(ep.Attributes["convert"].(string))
					if fnc.IsValid() {
						arguments := ep.Attributes
						arguments["vIn"] = vIn
						inputs := make([]reflect.Value, 1)

						inputs[0] = reflect.ValueOf(arguments)
						result = fnc.Call(inputs)[0].Float()
					} else {
						log.Printf("Converting function %s doesn't exist", ep.Attributes["convert"].(string))
					}
				} else {
					result = vIn * ep.Attributes["scale"].(float64)
				}
			}
			break
		}
	}

	return result
}

// OhmMeter function returning calculated value of resistance
func (ic *ICService) OhmMeter(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var vcc float64
	var result float64
	var scale float64

	arguments := common.Attributes{}
	result = -1.0
	scale = 1

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(common.Attributes)
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vcc"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vcc = arguments["vcc"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	// Calculating Ohm value
	if reference, ok := arguments["reference"]; ok {
		if reflect.TypeOf(reference).Kind() == reflect.Float64 {
			result = ((vcc/vIn - 1) * reference.(float64)) * scale
		}
	}
	return result

}

func (ic *ICService) ToLux(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var result float64
	var scale float64

	arguments := common.Attributes{}
	scale = 1
	result = -1.0

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(common.Attributes)
	} else {
		log.Printf("Invalid inputs --> %v", inputs)
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	// Calculating resulting value
	result = (vIn * (700 + math.Log10(vIn)*100)) * scale

	return result

}

//---------------------------------------------------------------------------
// template function returning calculated value of
func (ic *ICService) template(inputs interface{}) float64 {

	// function variables
	var vIn float64
	var result float64
	var scale float64

	arguments := common.Attributes{}
	scale = 1
	result = -1.0

	// Check if inputs parameter are Ok, if not returning "Error" value
	if reflect.TypeOf(inputs).Kind() == reflect.TypeOf(arguments).Kind() {
		arguments = inputs.(common.Attributes)
	} else {
		return result
	}

	// Checking inputs arguments and initializing function variables
	if input, ok := arguments["scale"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			scale = arguments["scale"].(float64)
		}
	}

	if input, ok := arguments["vIn"]; ok {
		if reflect.TypeOf(input).Kind() == reflect.Float64 {
			vIn = arguments["vIn"].(float64)
		} else {
			return result
		}
	} else {
		return result
	}

	/*
		if input, ok := arguments["vcc"]; ok {
			if reflect.TypeOf(input).Kind() == reflect.Float64 {
				vcc = arguments["vcc"].(float64)
			} else {
				return result
			}
		} else {
			return result
		}
	*/
	// Calculating resulting value
	result = vIn * scale
	return result

}
