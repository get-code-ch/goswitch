package model

import "strings"

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

func (node Node) SetFromInterface(m map[string]interface{}) Node {

	n := Node{}
	for key, value := range m {
		switch strings.ToLower(key) {
		case "id":
			n.Id = value.(string)
		case "type":
			n.Type = NodeType(value.(string))
		}
	}
	return n
}
