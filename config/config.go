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
}

type ConfCli struct {
	Controller ConfCommCtr   `json:"controller"`
	Interface  ConfInterface `json:"interface"`
}

type ConfCommCtr struct {
	ApiKey string          `json:"api_key"`
	Server string          `json:"server"`
	Port   string          `json:"port"`
	Ssl    bool            `json:"ssl"`
	Cert   ConfCertificate `json:"cert,omitempty"`
}

type ConfCertificate struct {
	SslKey  string `json:"ssl_key"`
	SslCert string `json:"ssl_cert,"`
}

type ConfInterface struct {
	Name string `json:"name"`
}
