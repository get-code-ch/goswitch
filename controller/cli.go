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
	"time"
)

type Cli struct {
	me       model.Node
	srv      model.Node
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

	cli.me = model.Node{Node: model.Cli, Id: cli.Name}
	cli.srv = model.Node{Node: model.Server, Id: "CommCtr"}

	/*
		d, err := json.Marshal(cli.me)
		if err != nil {
			log.Printf("Error marshaling cli: %v", err)
		}
	*/
	// Send Register message
	//msg := model.Message{Action: "Register", Data: string(d), Client: cli.me, Server: cli.srv}
	SendMessage(cli, model.Register, cli.me, nil)

	return cli
}

func (cli *Cli) Send(action model.Action, data interface{}, conn *websocket.Conn) {
	cli.conn.WriteJSON(model.Message{Action: action, Data: data, Client: cli.me, Server: cli.srv})
}

func (cli *Cli) Listen(channel chan int) {
	msg := new(model.Message)

	for {
		err := cli.conn.ReadJSON(&msg)
		if err != nil {
			if err.(*net.OpError).Err.(*os.SyscallError).Error() == "wsarecv: An existing connection was forcibly closed by the remote host." {
				log.Printf("Node closed by peer %v", err)
				cli.conn = nil
				for {
					time.Sleep(5 * time.Second)
					cli.conn, _, err = websocket.DefaultDialer.Dial(cli.url.String(), nil)
					if err == nil {
						log.Printf("Device reconnected\n")
						SendMessage(cli, model.Reconnect, cli.me, nil)
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

func (cli *Cli) Echo(msg string) {
	SendMessage(cli, model.Echo, msg, nil)
}

func (cli *Cli) Invoke(function model.Action, data interface{}) {
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(cli).MethodByName(string(function))
	if !fnc.IsValid() {
		SendMessage(cli, model.Error, fmt.Sprintf("Action %s not found", function), nil)
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

/*
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
*/
