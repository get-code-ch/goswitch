package model

import (
	"encoding/json"
	"fmt"
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
	BROADCAST   Action = "Broadcast"
	LIST        Action = "List"
	GETINFO     Action = "GetInfo"
	SENDINFO    Action = "ReceiveInfo"
	FAKE        Action = "Fake"
	SETGPIO     Action = "SetGPIO"
	GETGPIO     Action = "GetGPIO"
	GPIOSTATE   Action = "GPIOState"
	REJECT      Action = "Reject"
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

func (msg Message) SetFromInterface(data interface{}) Message {

	marshal, _ := json.Marshal(data)
	converted := Message{}
	json.Unmarshal(marshal, &converted)
	return converted
}
