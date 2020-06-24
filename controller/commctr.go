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
			log.Printf("ERROR starting server -> %v", err)
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
		log.Printf("ERROR handle wsListener --> %v", err)
		return
	}

	// REGISTER client (DEVICE, CLI or GUI)
	SendMessage(commCtr, commCtr.conn, model.REGISTER, nil)

	// Wait for message
	msg := new(model.Message)
	for {
		err := commCtr.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("ERROR reading wsListener --> %v", err)
			return
		}

		switch msg.Client.Node {
		case model.BROWSER, model.CLI:
			commCtr.Invoke(commCtr.clients[msg.Client.Id], msg.Action, msg.Data, msg.Client)
		case model.DEVICE:
			commCtr.Invoke(commCtr.devices[msg.Client.Id], msg.Action, msg.Data, msg.Client)
		}
	}
}
func (commCtr *CommandCenter) Send(conn *websocket.Conn, action model.Action, data interface{}) {
	conn.WriteJSON(model.Message{Action: action, Data: data})
}

func (commCtr *CommandCenter) Invoke(conn *websocket.Conn, function model.Action, args ...interface{}) {
	inputs := make([]reflect.Value, len(args)+1)
	inputs[0] = reflect.ValueOf(conn)
	for i := range args {
		inputs[i+1] = reflect.ValueOf(args[i])
	}

	fnc := reflect.ValueOf(commCtr).MethodByName(string(function))
	if !fnc.IsValid() {
		commCtr.conn.WriteJSON(model.Message{Action: "ERROR", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(conn *websocket.Conn, data interface{}, client model.Node) {
	d := data.(map[string]interface{})
	log.Printf("Id -> %s\n", d["Id"].(string))
	SendMessage(commCtr, conn, model.ACCEPT, nil)
}

func (commCtr *CommandCenter) Reconnect(conn *websocket.Conn, data interface{}, client model.Node) {
	d := data.(map[string]interface{})
	log.Printf("Id -> %s\n", d["Id"].(string))
	conn.WriteJSON(model.Message{Action: "Accept", Data: fmt.Sprintf("DEVICE %s Accepted", client.Id)})
}

func (commCtr *CommandCenter) Error(conn *websocket.Conn, data string) {
	log.Printf("ERROR function, data: %s", data)
}

func (commCtr *CommandCenter) Echo(conn *websocket.Conn, data string, client model.Node) {

	switch client.Node {
	case model.CLI, model.BROWSER:
		conn.WriteJSON(model.Message{Action: "Acknowledge", Data: fmt.Sprintf(`{"Message":"%s"}`, data)})
	case model.DEVICE:
		conn.WriteJSON(model.Message{Action: "Acknowledge", Data: fmt.Sprintf(`{"Message":"%s"}`, data)})
	}
}
