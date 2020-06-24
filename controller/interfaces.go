package controller

import (
	"github.com/gorilla/websocket"
	"goswitch/model"
)

type Controller interface {
	Listen(channel chan int)
	Send(action model.Action, data interface{}, conn *websocket.Conn)
	//Receive()
}

func SendMessage(controller Controller, action model.Action, data interface{}, conn *websocket.Conn) {
	controller.Send(action, data, conn)
}

func WaitMessages(controller Controller, channel chan int) {
	controller.Listen(channel)
}
