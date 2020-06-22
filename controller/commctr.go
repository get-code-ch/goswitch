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
	ssl      bool
}

func NewCommandCenter(conf *config.ConfCommCtr) *CommandCenter {

	commCtr := new(CommandCenter)
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

	return commCtr
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

		commCtr.Invoke(msg.Action, msg.Data)
	}
}

func (commCtr *CommandCenter) Invoke(function string, data string) {
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(data)
	fnc := reflect.ValueOf(commCtr).MethodByName(function)
	if !fnc.IsValid() {
		commCtr.conn.WriteJSON(model.Message{Action: "Error", Data: fmt.Sprintf("Action %s not found", function)})
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(data string) {
	device := new(Device)
	json.Unmarshal([]byte(data), &device)
	log.Printf("Device %s request register...", device.MacAddress)
	commCtr.conn.WriteJSON(model.Message{Action: "Accept", Data: "Accepted"})
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

func (commCtr *CommandCenter) Echo(data string) {
	echo := new(model.Echo)
	json.Unmarshal([]byte(data), &echo)

	commCtr.conn.WriteJSON(model.Message{Action: "Acknowledge", Data: fmt.Sprintf(`{"Message":"%s"}`, echo.Message)})
}
