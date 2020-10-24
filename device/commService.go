package main

import (
	"flag"
	"fmt"
	"github.com/get-code-ch/goswitch/common"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

type CommService struct {
	me     common.Node
	server common.Node
	url    url.URL
	header http.Header
	conn   *websocket.Conn
	mac    string
}

func (s *CommService) ConnectCommCenter(commCtr common.CommCenter, netAdapter string) {
	var err error
	var response *http.Response
	addr := flag.String("addr", fmt.Sprintf("%s:%s", commCtr.Server, commCtr.Port), "CommCenter http(s) address")
	flag.Parse()

	if commCtr.Ssl {
		s.url = url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	} else {
		s.url = url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	}

	// Adding origin in header, for server cross origin resource sharing (CORS) check
	s.header = http.Header{}
	s.header.Set("Origin", s.url.String())

	// Connecting CommandCenter server
	if s.conn, response, err = websocket.DefaultDialer.Dial(s.url.String(), s.header); err != nil {
		count := 0
		log.Printf("Dial Error %v\n", err)
		s.conn = nil

		for {
			time.Sleep(5 * time.Second)
			if s.conn, response, err = websocket.DefaultDialer.Dial(s.url.String(), s.header); err == nil {
				break
			} else {
				count++
				log.Printf("Dial Error %v (%d times)\n", err, count)
			}
		}
	}
	log.Printf("CommCenter connectected, (http status %d)", response.StatusCode)
	s.mac, _ = common.GetMacAddress(netAdapter)

	s.me = common.Node{Type: common.DEVICE, Id: s.mac}
	s.server = common.Node{Type: common.SERVER, Id: "CommCtr"}
}

func (s *CommService) Send(action common.Action, data interface{}) {
	s.conn.WriteJSON(common.Message{Action: action, Data: data, Client: s.me, Server: s.server})
}

func (s *CommService) Listen(channel chan int, device *Device) {
	msg := new(common.Message)
	count := 0

	for {
		if err := s.conn.ReadJSON(&msg); err != nil {
			log.Printf("Connection error -> %v", err)
			s.conn = nil
			for {
				time.Sleep(5 * time.Second)
				s.conn, _, err = websocket.DefaultDialer.Dial(s.url.String(), s.header)
				if err == nil {
					break
				} else {
					count++
					log.Printf("Connection error (%d) -> %v", count, err)
				}
			}
			continue
		}
		device.Invoke(msg.Action, msg.Data)
	}
}
