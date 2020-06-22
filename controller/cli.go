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
	"time"
)

type Cli struct {
	active   bool
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	url      url.URL
	ssl      bool
	Name     string
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
		//log.Fatal("dial error:", err)
		log.Printf("dial error: %v", err)
	}

	cli.Name, err = os.Hostname()
	if err != nil {
		cli.Name = "Unknown"
	}

	d, err := json.Marshal(cli)
	if err != nil {
		log.Printf("Error marshaling cli: %v", err)
	}
	msg := model.Message{Action: "Register", Data: string(d)}
	cli.conn.WriteJSON(msg)

	return cli
}

func (cli *Cli) Listen(channel chan int) {
	msg := new(model.Message)

	for {
		err := cli.conn.ReadJSON(&msg)
		if err != nil {
			if err.(*net.OpError).Err.(*os.SyscallError).Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				log.Printf("Connection closed by peer %v", err)
				cli.conn = nil
				for {
					time.Sleep(5 * time.Second)
					cli.conn, _, err = websocket.DefaultDialer.Dial(cli.url.String(), nil)
					if err == nil {
						log.Printf("Device reconnected\n")
						d, err := json.Marshal(cli)
						if err != nil {
							log.Printf("Error marshaling device: %v", err)
						}
						cli.conn.WriteJSON(model.Message{Action: "Reconnect", Data: string(d)})
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

		//log.Printf("Message received from %s (action: %s, msg:%s...)", cli.conn.RemoteAddr(), msg.Action, msg.Data)
		cli.Invoke(msg.Action, msg.Data)

	}

}

func (cli *Cli) Send(msg string) {
	cli.conn.WriteJSON(model.Message{Action: "Register", Data: msg})
}

func (cli *Cli) Echo(msg string) {
	cli.conn.WriteJSON(model.Message{Action: "Echo", Data: fmt.Sprintf(`{"Message":"%s"}`, msg)})
}

func (cli *Cli) Invoke(function string, data string) {
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(cli).MethodByName(function)
	if !fnc.IsValid() {
		cli.conn.WriteJSON(model.Message{Action: "Error", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (cli *Cli) Accept(data string) {
	log.Printf("Accept function, data: %s", data)
}

func (cli *Cli) Reconnect(data string) {
	log.Printf("Reconnect function, data: %s", data)
}

func (cli *Cli) Acknowledge(data string) {
	var err error
	ack := new(model.Acknowledge)
	err = json.Unmarshal([]byte(data), &ack)
	if err == nil {
		fmt.Printf("Acknowledge: %s\n", ack.Message)
	} else {
		fmt.Printf("Acknowledge Error: %s\n", err.Error())
	}
}
