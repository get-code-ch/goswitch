package config

import (
	"github.com/get-code-ch/mcp23008"
)

const defaultDeviceConfigFile = "./config/device.json"
const defaultCliConfigFile = "./config/cli.json"
const defaultControllerConfigFile = "./config/commctr.json"

type ConfDevice struct {
	Controller ConfCommCtr         `json:"controller"`
	Interface  ConfInterface       `json:"interface"`
	Modules    []mcp23008.Mcp23008 `json:"modules"`
	Switches   []I2cSwitch         `json:"switches"`
}

type ConfCli struct {
	Controller ConfCommCtr   `json:"controller"`
	Interface  ConfInterface `json:"interface"`
}

type ConfCommCtr struct {
	ApiKey     string          `json:"api_key"`
	Server     string          `json:"server"`
	Port       string          `json:"port"`
	Ssl        bool            `json:"ssl"`
	Cert       ConfCertificate `json:"cert,omitempty"`
	ClientRoot string          `json:"client_root"`
}

type I2cSwitch struct {
	Address int    `json:"address"`
	Gpio    int    `json:"gpio"`
	Name    string `json:"name"`
	State   int    `json:"state"`
}

type ConfCertificate struct {
	SslKey  string `json:"ssl_key"`
	SslCert string `json:"ssl_cert,"`
}

type ConfInterface struct {
	I2c  string `json:"i2c"`
	Name string `json:"name"`
}
