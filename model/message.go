package model

type Action string

const (
	Register    Action = "Register"
	Reconnect   Action = "Reconnect"
	Echo        Action = "Echo"
	Error       Action = "Error"
	Acknowledge Action = "Acknowledge"
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
