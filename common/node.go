package common

import "encoding/json"

type NodeType string

const (
	SERVER  NodeType = "server"
	DEVICE  NodeType = "device"
	CLI     NodeType = "cli"
	BROWSER NodeType = "browser"
)

type Node struct {
	Type NodeType
	Id   string
}

func (node Node) SetFromInterface(data interface{}) Node {

	marshal, _ := json.Marshal(data)
	converted := Node{}
	json.Unmarshal(marshal, &converted)
	return converted

}
