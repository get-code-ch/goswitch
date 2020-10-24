package main

import (
	"encoding/json"
	"fmt"
	"github.com/get-code-ch/goswitch/common"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

// NewDevice create a new device handler
func NewDevice(configFile string) *Device {

	// New device creation
	device := new(Device)

	// loading configuration
	device.config = loadConfig(configFile)

	// Connecting CommCenter
	device.commService = new(CommService)
	device.commService.ConnectCommCenter(device.config.CommCenter, device.config.Adapters.Network)

	// Initializing ICs
	device.icService = make(map[int]*ICService)
	for _, i := range device.config.ICs {
		device.icService[i.Address] = new(ICService)
		device.icService[i.Address].InitIC(i, device.config.Adapters.I2c, device.commService)
	}

	return device
}

// loadConfig read, parse and return configuration file
// If something is wrong log error message and abort program
func loadConfig(configFile string) Config {

	config := Config{}

	// If no Config file is provided we use default filepath
	if configFile == "" {
		configFile = defaultDeviceConfigFile
	}

	// Testing if Config file exist if not, return a fatal error
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Config file %s not exist\n", configFile)
		} else {
			log.Panicf("Something wrong with Config file %s -> %v\n", configFile, err)
		}
	}

	// Reading and parsing Config file
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading Config file --> %v", err)
	}

	// Parsing Config file
	if err := json.Unmarshal(buffer, &config); err != nil {
		log.Fatalf("Error parsing Config file --> %v", err)
	}

	return config
}

// TODO Create Stringer interface to return human readable Config content
func (device *Device) String() string {
	return fmt.Sprintf("Device configuration %v", device.config)
}

// Invoke launch dynamically an action (function) depending action
func (device *Device) Invoke(function common.Action, data interface{}) {

	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(device).MethodByName(string(function))
	if !fnc.IsValid() {
		device.commService.Send("ERROR", fmt.Sprintf("Action %s not found", function))
	} else {
		fnc.Call(inputs)
	}
}

// Register function register device on Communication Center
func (device *Device) Register(data interface{}) {
	arguments := make(map[string]interface{})
	arguments["client"] = device.commService.me
	arguments["api_key"] = device.config.CommCenter.ApiKey

	if !device.registered {
		device.commService.Send(common.REGISTER, arguments)
	} else {
		device.commService.Send(common.RECONNECT, device.commService.me)
	}
}

// Error function log an error message to the console (standard output)
func (device *Device) Error(data interface{}) {
	log.Printf("Error : %s\n> ", data.(string))
}

// Acknowledge function log acknowledgement to the console (standard output)
func (device *Device) Acknowledge(data interface{}) {
	log.Printf("Acknowledge received: %s\n", data.(string))
}

// Accept function log connection acceptation from CommCenter to the console (standard output)
// and sending list of device's ics
func (device *Device) Accept(data interface{}) {
	log.Printf("Connection accepted: %s\n", data.(string))
	device.commService.Send(common.ICS_LIST, device.config.ICs)
	//	device.GetAllGPIOState("")
}

// Reject function log rejected (unauthorized) connection from CommCenter to the console (standard output) and ending program
func (device *Device) Reject(data interface{}) {
	log.Fatalf("Connection rejected by Command Center by %s\n", data.(string))
}

// GetInfo function call IC GetInfo --> who sending info about IC Endpoints like temperature sensor or GPIO state
func (device *Device) GetInfo(data interface{}) {

	client := common.Node{}.SetFromInterface(data)
	for _, ic := range device.icService {
		ic.GetInfo(device, client)
	}

}

func (device *Device) SetGPIO(data interface{}) {

	request := data.(map[string]interface{})
	arguments := request["i2c"].(map[string]interface{})

	log.Printf("Received 'SetGPIO' Request --> %v", arguments)

}

func (device *Device) GetAllGPIOState(data interface{}) {

	log.Printf("Received 'GetAllGPIOState' Request --> %v", data)

}
