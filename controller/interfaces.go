package controller

import "goswitch/model"

type Controller interface {
	Listen(channel chan int)
	Send(action model.Action, data interface{})
	//Receive()
}

func SendMessage(controller Controller, action model.Action, data interface{}) {
	controller.Send(action, data)
}

func WaitMessages(controller Controller, channel chan int) {
	controller.Listen(channel)
}
