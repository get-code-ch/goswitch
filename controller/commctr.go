package controller

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"goswitch/config"
	"goswitch/model"
	"log"
	"net/http"
	"reflect"
)

type CommandCenter struct {
	active   bool
	upgrader websocket.Upgrader
	conn     *websocket.Conn
	devices  map[string]*websocket.Conn
	clients  map[string]*websocket.Conn
	ssl      bool
}

func NewCommandCenter() *CommandCenter {

	commCtr := new(CommandCenter)

	commCtr.devices = make(map[string]*websocket.Conn)
	commCtr.clients = make(map[string]*websocket.Conn)

	return commCtr
}

func (commCtr *CommandCenter) Listen(conf *config.ConfCommCtr, channel chan int) {
	addr := flag.String("addr", fmt.Sprintf("%s:%s", conf.Server, conf.Port), "https service address")
	flag.Parse()

	commCtr.ssl = conf.Ssl
	commCtr.upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	http.HandleFunc("/ws", commCtr.wsListener)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>goswitch server controller</h1>")
	})

	if commCtr.ssl {
		err := http.ListenAndServeTLS(*addr, conf.Cert.SslCert, conf.Cert.SslKey, nil)
		if err != nil {
			log.Printf("Error starting server -> %v", err)
		}
	} else {
		http.ListenAndServe(*addr, nil)
	}
	close(channel)
}

func (commCtr *CommandCenter) wsListener(w http.ResponseWriter, r *http.Request) {
	var err error

	// Init connection
	commCtr.conn, err = commCtr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error handle wsListener --> %v", err)
		return
	}

	// Wait for message
	msg := new(model.Message)
	for {
		err := commCtr.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading wsListener --> %v", err)
			return
		}

		commCtr.Invoke(msg.Action, msg.Data, msg.Client)
	}
}

func (commCtr *CommandCenter) Invoke(function model.Action, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	fnc := reflect.ValueOf(commCtr).MethodByName(string(function))
	if !fnc.IsValid() {
		commCtr.conn.WriteJSON(model.Message{Action: "Error", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(data string, buddy model.Node) {
	client := new(model.Node)
	json.Unmarshal([]byte(data), &client)
	log.Printf("Device %s request register...", client.Id)
	switch client.Node {
	case model.Device:
		commCtr.devices[client.Id] = commCtr.conn
		commCtr.devices[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Device %s Accepted", client.Id)})
	case model.Cli, model.Browser:
		commCtr.clients[client.Id] = commCtr.conn
		commCtr.clients[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Node %s Accepted", client.Id)})
	}
}

func (commCtr *CommandCenter) Reconnect(data string) {
	device := new(Device)
	json.Unmarshal([]byte(data), &device)
	log.Printf("Device %s reconnect...", device.MacAddress)
	commCtr.conn.WriteJSON(model.Message{Action: "Reconnect", Data: "Accepted"})
}

func (commCtr *CommandCenter) Error(data string) {
	log.Printf("Error function, data: %s", data)
}

func (commCtr *CommandCenter) Echo(data string, client model.Node) {

	switch client.Node {
	case model.Cli, model.Browser:
		commCtr.clients[client.Id].WriteJSON(model.Message{Action: "Acknowledge", Data: fmt.Sprintf(`{"Message":"%s"}`, data)})
	case model.Device:
		commCtr.devices[client.Id].WriteJSON(model.Message{Action: "Acknowledge", Data: fmt.Sprintf(`{"Message":"%s"}`, data)})
	}
}
