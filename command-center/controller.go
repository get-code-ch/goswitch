package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/get-code-ch/goswitch/common"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"
)

const defaultControllerConfigFile = "./config/commctr.json"

type CommandCenter struct {
	active            bool
	upgrader          websocket.Upgrader
	conn              *websocket.Conn
	devices           map[string]*websocket.Conn
	clients           map[string]*websocket.Conn
	me                common.Node
	ssl               bool
	cert              ConfCertificate
	server            string
	port              string
	clientRoot        string
	authorizedDevices []AuthorizedDevice
	corsOrigin        bool
	tmeConf           TmeConf
	mutex             sync.Mutex
}

func NewCommCtrConfig(configFile string) *ConfCommCtr {

	// New config creation
	c := new(ConfCommCtr)

	// If no config file is provided we use "hardcoded" default filepath
	if configFile == "" {
		configFile = defaultControllerConfigFile
	}

	// Testing if config file exist if not, return a fatal error
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Panic(fmt.Sprintf("Config file %s not exist\n", configFile))
		} else {
			log.Panic(fmt.Sprintf("Something wrong with config file %s -> %v\n", configFile, err))
		}
	}

	// Reading and parsing configuration file
	if buffer, err := ioutil.ReadFile(configFile); err != nil {
		log.Printf(fmt.Sprintf("Error reading config file --> %v", err))
		return nil
	} else {
		json.Unmarshal(buffer, c)
		return c
	}

}

// TODO Create Stringer interface to return human readable config content
func (c *ConfCommCtr) String() string {
	return fmt.Sprintf("Certificate Key %s\n", c.Cert.SslCert)
}

func NewCommandCenter(conf *ConfCommCtr) *CommandCenter {

	commCtr := new(CommandCenter)

	commCtr.devices = make(map[string]*websocket.Conn)
	commCtr.clients = make(map[string]*websocket.Conn)
	commCtr.ssl = conf.Ssl
	commCtr.cert.SslCert = conf.Cert.SslCert
	commCtr.cert.SslKey = conf.Cert.SslKey
	commCtr.server = conf.Server
	commCtr.port = conf.Port
	commCtr.clientRoot = conf.ClientRoot
	commCtr.authorizedDevices = []AuthorizedDevice{}

	for _, authDevice := range conf.AuthorizedDevices {
		if authDevice.Enabled {
			commCtr.authorizedDevices = append(commCtr.authorizedDevices, authDevice)
		}
	}

	commCtr.corsOrigin = conf.CorsOrigin
	commCtr.me = common.Node{Id: "CommCtr", Type: common.SERVER}

	// Get Telegram config
	commCtr.tmeConf = TmeConf{}
	if _, err := os.Stat(conf.TelegramConf); err == nil {
		if buffer, err := ioutil.ReadFile(conf.TelegramConf); err == nil {
			if err := json.Unmarshal(buffer, &commCtr.tmeConf); err != nil {
				log.Printf("Error getting Telegram configuration --> %v", err)
			}
		}
	}

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

	//fs := http.FileServer(http.Dir("./gsvue/dist/"))
	fs := http.FileServer(http.Dir(commCtr.clientRoot))
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
		re := regexp.MustCompile(`(?i)(?:[http|ws][s]?:\/\/)([^/]*)`)
		rHost := re.FindStringSubmatch(r.Header.Get("origin"))
		log.Printf("Origin: %s\n", r.Header.Get("origin"))
		log.Printf("Server -> %s:%s\n", commCtr.server, commCtr.port)
		if len(rHost) != 2 {
			return false
		}
		if commCtr.corsOrigin {
			if commCtr.server+":"+commCtr.port == rHost[1] {
				return true
			} else {
				return false
			}
		} else {
			return true
		}
	}

	conn, err = commCtr.upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Printf("ERROR handle serveWs --> %v", err)
		return
	}

	// Ask for client registration (DEVICE, CLI or GUI)
	commCtr.Send(conn, common.REGISTER, commCtr.me)
	// Wait for message
	msg := new(common.Message)
	err = conn.ReadJSON(&msg)
	commCtr.Invoke(conn, msg.Action, msg.Data, msg.Client)

	// Sending a acknowledge message to client every minutes
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		log.Printf("Timer started")
		for {
			select {
			case <-ticker.C:
				{
					commCtr.Send(conn, common.ACKNOWLEDGE, fmt.Sprintf("Ping %s", time.Now().Format("2006-01-02 15:04:05")))
				}
			}
		}
	}()

	for {
		err := conn.ReadJSON(&msg)
		if err != nil {
			ticker.Stop()
			ticker = nil
			log.Printf("ERROR reading serveWs --> %v", err)
			// Removing device or client/browser from list
			for key, value := range commCtr.clients {
				if value == conn {
					delete(commCtr.clients, key)
					conn.Close()
					return
				}
			}
			for key, value := range commCtr.devices {
				if value == conn {
					for dIdx, d := range commCtr.authorizedDevices {
						if d.MacAddr == key {
							commCtr.authorizedDevices[dIdx].IsOnline = false
							for _, c := range commCtr.clients {
								commCtr.Send(c, common.ACKNOWLEDGE, fmt.Sprintf("Device %s disconnected", key))
							}
						}
					}

					delete(commCtr.devices, key)
					for _, client := range commCtr.clients {
						commCtr.Send(client, common.ACKNOWLEDGE, fmt.Sprintf("Device %s disconnected", key))
						commCtr.List(client, nil, common.Node{})
					}
					conn.Close()
					return
				}
			}
			return
		}

		switch msg.Client.Type {
		case common.BROWSER, common.CLI:
			commCtr.Invoke(commCtr.clients[msg.Client.Id], msg.Action, msg.Data, msg.Client)
		case common.DEVICE:
			commCtr.Invoke(commCtr.devices[msg.Client.Id], msg.Action, msg.Data, msg.Client)
		}
	}
}
func (commCtr *CommandCenter) Send(conn *websocket.Conn, action common.Action, data interface{}) {
	commCtr.mutex.Lock()
	defer commCtr.mutex.Unlock()
	if conn != nil {
		if err := conn.WriteJSON(common.Message{Action: action, Data: data}); err != nil {
			log.Printf("Error sending message: %v", err)
			conn.Close()
			conn = nil
		}
	}
}

func (commCtr *CommandCenter) Invoke(conn *websocket.Conn, function common.Action, args ...interface{}) {
	inputs := make([]reflect.Value, len(args)+1)
	inputs[0] = reflect.ValueOf(conn)
	for i := range args {
		inputs[i+1] = reflect.ValueOf(args[i])
	}

	fnc := reflect.ValueOf(commCtr).MethodByName(string(function))
	if !fnc.IsValid() {
		//commCtr.conn.WriteJSON(model.Message{Action: "ERROR", Data: fmt.Sprintf("Action %s not found", function)})
		commCtr.Send(conn, common.ERROR, fmt.Sprintf("Action %s not found", function))
	} else {
		fnc.Call(inputs)
	}
}

func (commCtr *CommandCenter) Register(conn *websocket.Conn, data interface{}, client common.Node) {
	d := data.(map[string]interface{})
	key := d["api_key"]
	nIfg := d["client"].(map[string]interface{})

	node := new(common.Node)

	node.Id = nIfg["Id"].(string)
	node.Type = common.NodeType(nIfg["Type"].(string))

	switch node.Type {
	case common.BROWSER, common.CLI:
		// We accept only one connection from client/browser
		if _, exist := commCtr.clients[node.Id]; exist {
			commCtr.Send(commCtr.clients[node.Id], common.REJECT, "Session open in other browser window")
			commCtr.clients[node.Id].Close()
			delete(commCtr.clients, node.Id)
		}

		commCtr.clients[node.Id] = conn
		commCtr.Send(conn, common.ACCEPT, node.Id)
	case common.DEVICE:
		for dIdx, d := range commCtr.authorizedDevices {
			if d.MacAddr == node.Id && d.ApiKey == key {

				if _, exist := commCtr.devices[node.Id]; exist {
					commCtr.devices[node.Id].Close()
					delete(commCtr.devices, node.Id)
				}

				commCtr.authorizedDevices[dIdx].IsOnline = true
				commCtr.devices[node.Id] = conn
				for _, c := range commCtr.clients {
					commCtr.Send(c, common.ACKNOWLEDGE, fmt.Sprintf("Device %s connected", node.Id))
					commCtr.List(c, nil, common.Node{})
				}
				commCtr.Send(conn, common.ACCEPT, node.Id)
				return
			}
		}
		commCtr.Send(conn, common.REJECT, node.Id)
	}
}

func (commCtr *CommandCenter) Acknowledge(data interface{}) {
	log.Printf("Acknowledge received: %s", data.(string))
}

func (commCtr *CommandCenter) Echo(conn *websocket.Conn, data interface{}, client common.Node) {

	log.Printf("Echo request: %v", data)
	commCtr.Send(conn, common.ACKNOWLEDGE, data.(string))

}

func (commCtr *CommandCenter) Error(conn *websocket.Conn, data interface{}) {
	log.Printf("ERROR function, data: %v", data)
}

func (commCtr *CommandCenter) List(conn *websocket.Conn, data interface{}, client common.Node) {
	commCtr.Send(conn, common.LIST, commCtr.authorizedDevices)
}

func (commCtr *CommandCenter) Broadcast(conn *websocket.Conn, data interface{}, client common.Node) {
	msg := common.Message{}.SetFromInterface(data)

	switch msg.Client.Type {
	case common.BROWSER, common.CLI:
		for _, destConn := range commCtr.clients {
			if destConn != conn {
				commCtr.Send(destConn, msg.Action, msg.Data)
			}
		}
	case common.DEVICE:
		for _, destConn := range commCtr.devices {
			if destConn != conn {
				commCtr.Send(destConn, msg.Action, msg.Data)
			}
		}
	default:
		for _, destConn := range commCtr.clients {
			if destConn != conn {
				commCtr.Send(destConn, msg.Action, msg.Data)
			}
		}
		for _, destConn := range commCtr.devices {
			if destConn != conn {
				commCtr.Send(destConn, msg.Action, msg.Data)
			}
		}
	}
}

func (commCtr *CommandCenter) Relay(conn *websocket.Conn, data interface{}, client common.Node) {
	var destConn *websocket.Conn

	msg := common.Message{}.SetFromInterface(data)

	switch msg.Client.Type {
	case common.BROWSER, common.CLI:
		destConn = commCtr.clients[msg.Client.Id]
	case common.DEVICE:
		destConn = commCtr.devices[msg.Client.Id]
	default:
		destConn = nil
	}

	if destConn != nil {
		commCtr.Send(destConn, msg.Action, msg.Data)
	} else {
		//SendMessage(commCtr, conn, common.ERROR, fmt.Sprintf("Device %s not found", msg.Client.Id))
		log.Printf("Device %s not found", msg.Client.Id)
	}
}

func (commCtr *CommandCenter) ToTelegram(conn *websocket.Conn, data interface{}, client common.Node) {

	if commCtr.tmeConf == (TmeConf{}) {
		log.Printf("Telegram bot not configured, message ignored")
		return
	}
	msg := common.Message{}.SetFromInterface(data)

	tmeUrl := url.URL{Host: "api.telegram.org", Scheme: "https", Path: "/" + commCtr.tmeConf.BotId + "/sendMessage"}
	tmeBody, _ := json.Marshal(TmeMessage{ChatId: commCtr.tmeConf.ChatId, DisableNotification: false, Text: msg.Data.(string)})

	if request, err := http.NewRequest("POST", tmeUrl.String(), bytes.NewBuffer(tmeBody)); err == nil {
		request.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		if response, err := client.Do(request); err != nil {
			log.Printf("Error sending message to Telegram --> %v\n", err)
		} else {
			log.Printf("Message sent to Telegram with status %d\n", response.StatusCode)
		}
	} else {
		log.Printf("Error creation http Request --> %v\n", err)
	}

}

func (commCtr *CommandCenter) ICsList(conn *websocket.Conn, data interface{}, client common.Node) {

	log.Printf("ICsList, data: %v", data)
}
