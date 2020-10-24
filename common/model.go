package common

type CommCenter struct {
	ApiKey string `json:"api_key"`
	Server string `json:"server"`
	Port   string `json:"port"`
	Ssl    bool   `json:"ssl"`
}

type Adapters struct {
	I2c     string `json:"i2c"`
	Network string `json:"network"`
}

type IC struct {
	Address     int        `json:"address"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        IcRef      `json:"type"`
	Endpoints   []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	Attributes Attributes `json:"attributes"`
}

type Attributes map[string]interface{}

type IcRef string

const (
	MCP23008 IcRef = "mcp23008"
	ADS1115  IcRef = "ads1115"
)
