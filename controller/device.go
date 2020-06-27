package controller

import (
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
	me       model.Node
	srv      model.Node
	active   bool
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	url      url.URL
	ssl      bool
	Name     string
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

	device.me = model.Node{Type: model.DEVICE, Id: device.Name}
	device.srv = model.Node{Type: model.SERVER, Id: "CommCtr"}

	return device
}

func (device *Device) setMacAddress(name string) {
	interfaces, err := net.Interfaces()
	device.Name = RandomString(8)
	//device.Name = ""
	if err == nil {
		for _, i := range interfaces {
			if strings.ToLower(i.Name) == strings.ToLower(name) {
				device.Name = i.HardwareAddr.String()
				break
			}
		}
	}
}

func (device *Device) Send(conn *websocket.Conn, action model.Action, data interface{}) {
	device.conn.WriteJSON(model.Message{Action: action, Data: data, Client: device.me, Server: device.srv})
}

func (device *Device) Listen(channel chan int) {
	msg := new(model.Message)

	for {
		err := device.conn.ReadJSON(&msg)
		if err != nil {
			if err.(*net.OpError).Err.(*os.SyscallError).Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				log.Printf("Type closed by peer %v", err)
				device.conn = nil
				for {
					time.Sleep(5 * time.Second)
					device.conn, _, err = websocket.DefaultDialer.Dial(device.url.String(), nil)
					if err == nil {
						log.Printf("DEVICE reconnected\n")
						SendMessage(device, nil, model.RECONNECT, device.me)
						break
					}
				}
				continue
			} else {
				log.Printf("ERROR reading websocket --> %v", err)
				close(channel)
				return
			}
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
	SendMessage(device, nil, model.REGISTER, device.me)
}

func (device *Device) Acknowledge(data interface{}) {
	log.Printf("Acknowledge received: %s\n", data.(string))
}

func (device *Device) Accept(data interface{}) {
	log.Printf("Connection accepted: %s\n", data.(string))
}

func (device *Device) GetInfo(data interface{}) {
	client := model.Node{}.SetFromInterface(data.(map[string]interface{}))

	hostName, _ := os.Hostname()
	info := model.Message{Action: model.SENDINFO, Data: fmt.Sprintf("Device Hostname is -> %s", hostName), Client: client}

	//msg := model.Message{Action: nil, Client: itf.Client, Data: info}
	SendMessage(device, nil, model.RELAY, info)
}
