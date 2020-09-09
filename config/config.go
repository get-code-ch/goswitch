// Package config provide tools to load configuration of devices and command center.
package config

import (
	"github.com/get-code-ch/ads1115"
	"github.com/get-code-ch/mcp23008/v3"
)

const defaultDeviceConfigFile = "./config/device.json"
const defaultCliConfigFile = "./config/cli.json"
const defaultControllerConfigFile = "./config/commctr.json"

type ConfDevice struct {
	Controller ConfCommCtr               `json:"controller"`
	Interface  ConfInterface             `json:"interface"`
	Name       string                    `json:"name"`
	Modules    map[int]mcp23008.Mcp23008 `json:"modules"`
	Switches   []I2cSwitch               `json:"switches"`
}

type ConfCli struct {
	Controller ConfCommCtr   `json:"controller"`
	Interface  ConfInterface `json:"interface"`
}

type ConfCommCtr struct {
	ApiKey            string             `json:"api_key"`
	Server            string             `json:"server"`
	Port              string             `json:"port"`
	Ssl               bool               `json:"ssl"`
	Cert              ConfCertificate    `json:"cert,omitempty"`
	ClientRoot        string             `json:"client_root"`
	AuthorizedDevices []AuthorizedDevice `json:"authorized_devices"`
	CorsOrigin        bool               `json:"cors_origin"`
}

type AuthorizedDevice struct {
	ApiKey   string `json:"api_key"`
	Name     string `json:"name"`
	MacAddr  string `json:"mac_addr"`
	IsOnline bool   `json:"is_online"`
	Enabled  bool   `json:"enabled"`
}

type I2cSwitch struct {
	MacAddr string `json:"mac_addr"`
	Address int    `json:"address"`
	Gpio    int    `json:"gpio"`
	Name    string `json:"name"`
	State   int    `json:"state"`
}

type I2cADC struct {
	MacAddr string      `json:"mac_addr"`
	Address int         `json:"address"`
	Name    string      `json:"name"`
	AIN     ads1115.Mux `json:"mux"`
}

type ConfCertificate struct {
	SslKey  string `json:"ssl_key"`
	SslCert string `json:"ssl_cert,"`
}

type ConfInterface struct {
	I2c  string `json:"i2c"`
	Name string `json:"name"`
}
