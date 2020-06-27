package controller

import (
	"github.com/gorilla/websocket"
	"goswitch/model"
	"math/rand"
	"strings"
	"time"
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

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	if length < 1 {
		length = 10
	}

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
