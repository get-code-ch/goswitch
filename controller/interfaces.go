package controller

import (
	"github.com/gorilla/websocket"
	"goswitch/model"
)

type Controller interface {
	Listen(channel chan int)
	Send(conn *websocket.Conn, action model.Action, data interface{})
	//Receive()
}

func SendMessage(controller Controller, conn *websocket.Conn, action model.Action, data interface{}) {
	controller.Send(conn, action, data)
}

func WaitMessages(controller Controller, channel chan int) {
	controller.Listen(channel)
}
