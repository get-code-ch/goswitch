package model

type Action string

const (
	REGISTER    Action = "Register"
	RECONNECT   Action = "Reconnect"
	ECHO        Action = "Echo"
	ERROR       Action = "Error"
	ACKNOWLEDGE Action = "Acknowledge"
	ACCEPT      Action = "Accept"
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
