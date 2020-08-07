package controller

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/get-code-ch/goswitch/config"
	"github.com/get-code-ch/goswitch/model"
	"github.com/get-code-ch/mcp23008/v2"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
)

type Device struct {
	me         model.Node
	apiKey     string
	srv        model.Node
	registered bool
	upgrader   websocket.Upgrader
	conn       *websocket.Conn
	url        url.URL
	ssl        bool
	MacAddr    string
	I2c        string
	I2cMode    model.I2cMode
	Modules    map[int]mcp23008.Mcp23008 `json:"modules"`
	Switches   []config.I2cSwitch        `json:"switches"`
}

func (device Device) SetFromInterface(data interface{}) Device {
	marshal, _ := json.Marshal(data)
	converted := Device{}
	json.Unmarshal(marshal, &converted)
	return converted
}

type DeviceInfo struct {
	Hostname string `json:"hostname"`
	Device   Device `json:"device"`
}

func (deviceInfo DeviceInfo) SetFromInterface(data interface{}) DeviceInfo {
	marshal, _ := json.Marshal(data)
	converted := DeviceInfo{}
	json.Unmarshal(marshal, &converted)
	return converted
}

func NewDevice(conf *config.ConfDevice) *Device {
	var err error

	device := new(Device)
	addr := flag.String("addr", fmt.Sprintf("%s:%s", conf.Controller.Server, conf.Controller.Port), "https service address")
	flag.Parse()

	if conf.Controller.Ssl {
		device.url = url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	} else {
		device.url = url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	}
	device.conn, _, err = websocket.DefaultDialer.Dial(device.url.String(), nil)
	if err != nil {
		count := 0
		log.Printf("Dial error -> %v", err)

		device.conn = nil
		for {
			time.Sleep(5 * time.Second)
			device.conn, _, err = websocket.DefaultDialer.Dial(device.url.String(), nil)
			if err == nil {
				break
			} else {
				count++
				log.Printf("Dial error (%d) -> %v", count, err)
			}
		}
	}

	device.setMacAddress(conf.Interface.Name)

	device.me = model.Node{Type: model.DEVICE, Id: device.MacAddr}
	device.apiKey = conf.Controller.ApiKey
	device.srv = model.Node{Type: model.SERVER, Id: "CommCtr"}
	device.I2c = conf.Interface.I2c
	device.registered = false

	return device
}

func (device *Device) setMacAddress(name string) {
	device.MacAddr = ""

	// Getting list of network interfaces
	interfaces, err := net.Interfaces()
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].HardwareAddr.String() < interfaces[j].HardwareAddr.String()
	})

	// Try to find interface matching with name
	if err == nil {
		for _, i := range interfaces {
			if strings.ToLower(i.Name) == strings.ToLower(name) {
				device.MacAddr = i.HardwareAddr.String()
				break
			}
		}
	}

	// If no device match with name we get mac address of first active interface (except loopback)
	if device.MacAddr == "" && err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp == net.FlagUp &&
				i.Flags&net.FlagBroadcast == net.FlagBroadcast &&
				i.Flags&net.FlagLoopback != net.FlagLoopback {
				device.MacAddr = i.HardwareAddr.String()
				break
			}
		}
	}

	// If no device was found setting mac address with a random string
	if device.MacAddr == "" {
		device.MacAddr = RandomString(8)
	}
}

func (device *Device) Send(conn *websocket.Conn, action model.Action, data interface{}) {
	device.conn.WriteJSON(model.Message{Action: action, Data: data, Client: device.me, Server: device.srv})
}

func (device *Device) Listen(channel chan int) {
	msg := new(model.Message)
	count := 0

	for {
		err := device.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Connection error -> %v", err)
			device.conn = nil
			for {
				time.Sleep(5 * time.Second)
				device.conn, _, err = websocket.DefaultDialer.Dial(device.url.String(), nil)
				if err == nil {
					break
				} else {
					count++
					log.Printf("Connection error (%d) -> %v", count, err)
				}
			}
			continue
		}
		device.Invoke(msg.Action, msg.Data)

	}
}

func (device *Device) Invoke(function model.Action, data interface{}) {

	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(device).MethodByName(string(function))
	if !fnc.IsValid() {
		device.conn.WriteJSON(model.Message{Action: "ERROR", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (device *Device) Register(data interface{}) {
	arguments := make(map[string]interface{})
	arguments["client"] = device.me
	arguments["api_key"] = device.apiKey

	if !device.registered {
		SendMessage(device, nil, model.REGISTER, arguments)
	} else {
		SendMessage(device, nil, model.RECONNECT, device.me)
	}
}
func (device *Device) Error(data interface{}) {
	log.Printf("Error : %s\n> ", data.(string))
}

func (device *Device) Acknowledge(data interface{}) {
	log.Printf("Acknowledge received: %s\n", data.(string))
}

func (device *Device) Accept(data interface{}) {
	log.Printf("Connection accepted: %s\n", data.(string))
	device.GetAllGPIOState("")
}

func (device *Device) Reject(data interface{}) {
	log.Fatalf("Connection rejected by Command Center by %s\n", data.(string))
}

func (device *Device) GetInfo(data interface{}) {
	hostName, _ := os.Hostname()

	client := model.Node{}.SetFromInterface(data)
	deviceInfo := DeviceInfo{Hostname: hostName, Device: *device}

	//info := model.Message{Action: model.SENDINFO, Data: deviceInfo, Client: model.Node{Id: "", Type: model.CLI}, Server: device.me}
	info := model.Message{Action: model.SENDINFO, Data: deviceInfo, Client: client, Server: device.me}

	SendMessage(device, nil, model.RELAY, info)
	device.GetAllGPIOState("")
}
