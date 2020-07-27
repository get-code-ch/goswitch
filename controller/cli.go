package controller

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"goswitch/config"
	"goswitch/model"
	"log"
	"net/url"
	"os"
	"reflect"
	"time"
)

type Cli struct {
	me         model.Node
	srv        model.Node
	registered bool
	upgrader   websocket.Upgrader
	conn       *websocket.Conn
	url        url.URL
	ssl        bool
	Name       string
	Devices    map[string]DeviceInfo
}

func NewCli(conf *config.ConfCli) *Cli {
	var err error

	cli := new(Cli)
	addr := flag.String("addr", fmt.Sprintf("%s:%s", conf.Controller.Server, conf.Controller.Port), "https service address")
	flag.Parse()

	if conf.Controller.Ssl {
		cli.url = url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	} else {
		cli.url = url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	}
	cli.conn, _, err = websocket.DefaultDialer.Dial(cli.url.String(), nil)
	if err != nil {
		count := 0
		log.Printf("Dial error -> %v", err)

		cli.conn = nil
		for {
			time.Sleep(5 * time.Second)
			cli.conn, _, err = websocket.DefaultDialer.Dial(cli.url.String(), nil)
			if err == nil {
				break
			} else {
				count++
				log.Printf("Dial error (%d) -> %v", count, err)
			}
		}
	}

	cli.Name, err = os.Hostname()
	if err != nil {
		cli.Name = "Unknown"
	}
	cli.Name += "-" + RandomString(8)

	cli.me = model.Node{Type: model.CLI, Id: cli.Name}
	cli.srv = model.Node{Type: model.SERVER, Id: "CommCtr"}
	cli.registered = false

	cli.Devices = make(map[string]DeviceInfo)

	return cli
}

func (cli *Cli) Send(conn *websocket.Conn, action model.Action, data interface{}) {
	cli.conn.WriteJSON(model.Message{Action: action, Data: data, Client: cli.me, Server: cli.srv})
}

func (cli *Cli) Listen(channel chan int) {
	msg := new(model.Message)
	count := 0

	for {
		err := cli.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Connection error -> %v", err)
			cli.conn = nil
			for {
				time.Sleep(5 * time.Second)
				cli.conn, _, err = websocket.DefaultDialer.Dial(cli.url.String(), nil)
				if err == nil {
					break
				} else {
					count++
					log.Printf("Connection error (%d) -> %v", count, err)
				}
			}
			continue
		}
		cli.Invoke(msg.Action, msg.Data)
	}

}

func (cli *Cli) Invoke(function model.Action, data interface{}) {
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(cli).MethodByName(string(function))
	if !fnc.IsValid() {
		SendMessage(cli, nil, model.ERROR, fmt.Sprintf("Action %s not found", function))
	} else {
		fnc.Call(inputs)
	}
}

func (cli *Cli) Register(data interface{}) {
	if !cli.registered {
		SendMessage(cli, nil, model.REGISTER, cli.me)
	} else {
		SendMessage(cli, nil, model.RECONNECT, cli.me)
	}
}

func (cli *Cli) Error(data interface{}) {
	log.Printf("Error : %s\n> ", data.(string))
}

func (cli *Cli) Acknowledge(data interface{}) {
	fmt.Printf("Acknowledge received: %s\n> ", data.(string))
}

func (cli *Cli) Accept(data interface{}) {
	fmt.Printf("Connection accepted: %s\n> ", data.(string))
}

func (cli *Cli) ReceiveInfo(data interface{}) {

	deviceInfo := DeviceInfo{}.SetFromInterface(data)

	// Insert or update devices info
	cli.Devices[deviceInfo.Device.me.Id] = deviceInfo

	fmt.Printf("\n")
	for _, swc := range deviceInfo.Device.Switches {
		fmt.Printf("Name -> %s (%d), GPIO %d - state %d ", swc.Name, swc.Address, swc.Gpio, swc.State)
		fmt.Printf("\n")
	}

	fmt.Printf("\n> ")
}

func (cli *Cli) Echo(data interface{}) {
	SendMessage(cli, nil, model.ACKNOWLEDGE, data.(string))
}

func (cli *Cli) List(data interface{}) {

	itf := data.([]interface{})
	deviceLst := make([]string, len(itf))
	for idx, value := range itf {
		deviceLst[idx] = value.(string)
	}

	if len(deviceLst) == 1 && deviceLst[0] == "" {
		deviceLst = nil
	}

	for i, device := range deviceLst {
		fmt.Printf("Device %d - %s\n", i, device)
	}
	fmt.Printf("> ")

}
