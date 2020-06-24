package controller

import (
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
	cert     config.ConfCertificate
	server   string
	port     string
}

func NewCommandCenter(conf *config.ConfCommCtr) *CommandCenter {

	commCtr := new(CommandCenter)

	commCtr.devices = make(map[string]*websocket.Conn)
	commCtr.clients = make(map[string]*websocket.Conn)
	commCtr.ssl = conf.Ssl
	commCtr.cert.SslCert = conf.Cert.SslCert
	commCtr.cert.SslKey = conf.Cert.SslKey
	commCtr.server = conf.Server
	commCtr.port = conf.Port

	return commCtr
}

func (commCtr *CommandCenter) Listen(channel chan int) {
	addr := flag.String("addr", fmt.Sprintf("%s:%s", commCtr.server, commCtr.port), "https service address")
	flag.Parse()

	commCtr.upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	http.HandleFunc("/ws", commCtr.wsListener)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>goswitch server controller</h1>")
	})

	if commCtr.ssl {
		err := http.ListenAndServeTLS(*addr, commCtr.cert.SslCert, commCtr.cert.SslKey, nil)
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
func (commCtr *CommandCenter) Send(action model.Action, data interface{}, conn *websocket.Conn) {
	conn.WriteJSON(model.Message{Action: action, Data: data})
}

func (commCtr *CommandCenter) Invoke(function model.Action, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	fnc := reflect.ValueOf(commCtr).MethodByName(string(function))
	if !fnc.IsValid() {
		commCtr.conn.WriteJSON(model.Message{Action: "Error", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(data interface{}, client model.Node, conn *websocket.Conn) {
	d := data.(map[string]interface{})
	log.Printf("Id -> %s\n", d["Id"].(string))
	switch client.Node {
	case model.Device:
		commCtr.devices[client.Id] = conn
		commCtr.devices[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Device %s Accepted", client.Id)})
	case model.Cli, model.Browser:
		commCtr.clients[client.Id] = conn
		commCtr.clients[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Node %s Accepted", client.Id)})
	}
}

func (commCtr *CommandCenter) Reconnect(data interface{}, client model.Node) {
	d := data.(map[string]interface{})
	log.Printf("Id -> %s\n", d["Id"].(string))
	switch client.Node {
	case model.Device:
		commCtr.devices[client.Id] = commCtr.conn
		commCtr.devices[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Device %s Accepted", client.Id)})
	case model.Cli, model.Browser:
		commCtr.clients[client.Id] = commCtr.conn
		commCtr.clients[client.Id].WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("Node %s Accepted", client.Id)})
	}
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
