package controller

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"goswitch/config"
	"goswitch/model"
	"log"
	"net"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
)

type Device struct {
	active     bool
	upgrader   websocket.Upgrader
	conn       *websocket.Conn
	url        url.URL
	ssl        bool
	MacAddress string `json:"mac_address"`
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
		//log.Fatal("dial error:", err)
		log.Printf("dial error: %v", err)
	}

	device.setMacAddress(conf.Interface.Name)

	d, err := json.Marshal(model.Node{Node: model.Device, Id: device.MacAddress})
	if err != nil {
		log.Printf("Error marshaling device: %v", err)
	}
	msg := model.Message{Action: "Register", Data: string(d)}
	device.conn.WriteJSON(msg)

	return device
}

func (device *Device) setMacAddress(name string) {
	interfaces, err := net.Interfaces()
	device.MacAddress = ""
	if err == nil {
		for _, i := range interfaces {
			if strings.ToLower(i.Name) == strings.ToLower(name) {
				device.MacAddress = i.HardwareAddr.String()
				break
			}
		}
	}
}

func (device *Device) Listen(channel chan int) {
	msg := new(model.Message)

	for {
		err := device.conn.ReadJSON(&msg)
		if err != nil {
			if err.(*net.OpError).Err.(*os.SyscallError).Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				log.Printf("Node closed by peer %v", err)
				device.conn = nil
				for {
					time.Sleep(5 * time.Second)
					device.conn, _, err = websocket.DefaultDialer.Dial(device.url.String(), nil)
					if err == nil {
						log.Printf("Device reconnected\n")
						d, err := json.Marshal(model.Node{Node: model.Device, Id: device.MacAddress})
						if err != nil {
							log.Printf("Error marshaling device: %v", err)
						}
						device.conn.WriteJSON(model.Message{Action: "Reconnect", Data: string(d)})
						break
					}
				}
				continue
			} else {
				log.Printf("Error reading websocket --> %v", err)
				close(channel)
				return
			}
		}

		log.Printf("Message received from %s (action: %s, msg:%s...)", device.conn.RemoteAddr(), msg.Action, msg.Data)
		device.Invoke(msg.Action, msg.Data)

	}

}

func (device *Device) Send(msg string) {
	device.conn.WriteJSON(model.Message{Action: "Register", Data: msg})
}

func (device *Device) Invoke(function model.Action, data interface{}) {
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(device).MethodByName(string(function))
	if !fnc.IsValid() {
		device.conn.WriteJSON(model.Message{Action: "Error", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (device *Device) Accept(data string) {
	log.Printf("Accept function, data: %s", data)
}

func (device *Device) Reconnect(data string) {
	log.Printf("Reconnect function, data: %s", data)
}
