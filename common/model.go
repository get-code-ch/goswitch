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
	Id              string     `json:"id"`
	Name            string     `json:"name"`
	RefreshInterval int        `json:"refresh_interval"`
	Telegram        Telegram   `json:"telegram,omitempty"`
	Attributes      Attributes `json:"attributes"`
}

type Telegram struct {
	Notification TmeNotification `json:"notification"`
	Max          float64         `json:"max,omitempty"`
	Min          float64         `json:"min,omitempty"`
}

type TmeNotification string

const (
	ONCHANGE TmeNotification = "onchange"
	VALUE    TmeNotification = "value"
	YES      TmeNotification = "yes"
	NO       TmeNotification = "no"
)

type Attributes map[string]interface{}

type IcRef string

const (
	MCP23008 IcRef = "mcp23008"
	ADS1115  IcRef = "ads1115"
)
