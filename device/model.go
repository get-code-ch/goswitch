package main

import "github.com/get-code-ch/goswitch/common"

const defaultDeviceConfigFile = "./config/device.json"

type Device struct {
	config      Config
	commService *CommService
	icService   map[int]*ICService
	registered  bool
}

type Config struct {
	Name       string            `json:"name"`
	CommCenter common.CommCenter `json:"comm_center"`
	Adapters   common.Adapters   `json:"adapters"`
	ICs        []common.IC       `json:"ics"`
}
