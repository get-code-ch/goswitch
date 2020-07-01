package model

import (
	"fmt"
	"strings"
)

type Action string

const (
	REGISTER    Action = "Register"
	RECONNECT   Action = "Reconnect"
	ECHO        Action = "Echo"
	ERROR       Action = "Error"
	ACKNOWLEDGE Action = "Acknowledge"
	ACCEPT      Action = "Accept"
	RELAY       Action = "Relay"
	LIST        Action = "List"
	GETINFO     Action = "GetInfo"
	SENDINFO    Action = "ReceiveInfo"
)

type Message struct {
	Client Node        `json:"client"`
	Server Node        `json:"server"`
	Action Action      `json:"action"`
	Data   interface{} `json:"data"`
}

func (a Action) String() string {
	return string(a)
}

func (msg Message) String() string {
	return fmt.Sprintf("{\nClient: {Type: %s, Id: %s}\n}", msg.Client.Id, msg.Client.Type)
}

func (msg Message) SetFromInterface(m map[string]interface{}) Message {

	msgStruct := Message{}

	for key, value := range m {
		switch strings.ToLower(key) {
		case "client":
			msgStruct.Client = Node{}.SetFromInterface(value.(map[string]interface{}))
		case "action":
			msgStruct.Action = Action(value.(string))
		case "data":
			msgStruct.Data = value
		case "server":
			msgStruct.Server = Node{}.SetFromInterface(value.(map[string]interface{}))
		}
	}
	return msgStruct
}
