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
	me       model.Node
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

	commCtr.me = model.Node{Id: "CommCtr", Type: model.SERVER}

	return commCtr
}

func (commCtr *CommandCenter) Listen(channel chan int) {
	addr := flag.String("addr", fmt.Sprintf("%s:%s", commCtr.server, commCtr.port), "https service address")
	flag.Parse()

	commCtr.upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	http.HandleFunc("/ws", commCtr.serveWs)

	http.HandleFunc("/title", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<h1>goswitch server controller</h1>")
	})

	fs := http.FileServer(http.Dir("./gsvue/dist/"))
	http.Handle("/", fs)

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

func (commCtr *CommandCenter) serveWs(w http.ResponseWriter, r *http.Request) {
	var err error
	var conn *websocket.Conn

	// Init connection

	header := http.Header{}

	// For development we allow CORS
	commCtr.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err = commCtr.upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Printf("ERROR handle serveWs --> %v", err)
		return
	}

	// Ask for client registration (DEVICE, CLI or GUI)
	SendMessage(commCtr, conn, model.REGISTER, commCtr.me)
	// Wait for message
	msg := new(model.Message)
	err = conn.ReadJSON(&msg)
	commCtr.Invoke(conn, msg.Action, msg.Data, msg.Client)

	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("ERROR reading serveWs --> %v", err)
			return
		}

		switch msg.Client.Type {
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
		//commCtr.conn.WriteJSON(model.Message{Action: "ERROR", Data: fmt.Sprintf("Action %s not found", function)})
		SendMessage(commCtr, conn, model.ERROR, fmt.Sprintf("Action %s not found", function))
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(conn *websocket.Conn, data interface{}, client model.Node) {
	d := data.(map[string]interface{})
	node := new(model.Node)
	node.Id = d["Id"].(string)
	node.Type = model.NodeType(d["Type"].(string))

	switch node.Type {
	case model.BROWSER, model.CLI:
		commCtr.clients[node.Id] = conn
	case model.DEVICE:
		commCtr.devices[node.Id] = conn
	}
	SendMessage(commCtr, conn, model.ACCEPT, node.Id)
}

func (commCtr *CommandCenter) Reconnect(conn *websocket.Conn, data interface{}, client model.Node) {
	d := data.(map[string]interface{})
	node := new(model.Node)
	node.Id = d["Id"].(string)
	node.Type = model.NodeType(d["Type"].(string))

	switch node.Type {
	case model.BROWSER, model.CLI:
		commCtr.clients[node.Id] = conn
	case model.DEVICE:
		commCtr.devices[node.Id] = conn
	}
	SendMessage(commCtr, conn, model.ACCEPT, node.Id)
}

func (commCtr *CommandCenter) Error(conn *websocket.Conn, data interface{}) {
	log.Printf("ERROR function, data: %v", data)
}

func (commCtr *CommandCenter) Echo(conn *websocket.Conn, data interface{}, client model.Node) {

	log.Printf("Echo request: %v", data)
	SendMessage(commCtr, conn, model.ACKNOWLEDGE, data.(string))

}

func (commCtr *CommandCenter) List(conn *websocket.Conn, data interface{}, client model.Node) {
	var deviceLst []string

	for key, _ := range commCtr.devices {
		deviceLst = append(deviceLst, key)
	}
	if deviceLst == nil {
		deviceLst = make([]string, 1)
	}
	SendMessage(commCtr, conn, model.LIST, deviceLst)

}

func (commCtr *CommandCenter) Acknowledge(data interface{}) {
	log.Printf("Acknowledge received: %s", data.(string))
}

func (commCtr *CommandCenter) Relay(conn *websocket.Conn, data interface{}, client model.Node) {
	var destConn *websocket.Conn

	msg := model.Message{}.SetFromInterface(data)

	switch msg.Client.Type {
	case model.BROWSER, model.CLI:
		destConn = commCtr.clients[msg.Client.Id]
	case model.DEVICE:
		destConn = commCtr.devices[msg.Client.Id]
	default:
		destConn = nil
	}

	if destConn != nil {
		SendMessage(commCtr, destConn, msg.Action, msg.Data)
	} else {
		SendMessage(commCtr, conn, model.ERROR, fmt.Sprintf("Device %s not found", msg.Client.Id))
	}
}

func (commCtr *CommandCenter) Broadcast(conn *websocket.Conn, data interface{}, client model.Node) {
	msg := model.Message{}.SetFromInterface(data)

	switch msg.Client.Type {
	case model.BROWSER, model.CLI:
		for _, destConn := range commCtr.clients {
			if destConn != conn {
				SendMessage(commCtr, destConn, msg.Action, msg.Data)
			}
		}
	case model.DEVICE:
		for _, destConn := range commCtr.devices {
			if destConn != conn {
				SendMessage(commCtr, destConn, msg.Action, msg.Data)
			}
		}
	default:
		for _, destConn := range commCtr.clients {
			if destConn != conn {
				SendMessage(commCtr, destConn, msg.Action, msg.Data)
			}
		}
		for _, destConn := range commCtr.devices {
			if destConn != conn {
				SendMessage(commCtr, destConn, msg.Action, msg.Data)
			}
		}
	}

}
