package model

type NodeType string

const (
	SERVER  NodeType = "server"
	DEVICE  NodeType = "device"
	CLI     NodeType = "cli"
	BROWSER NodeType = "browser"
)

type Node struct {
	Node NodeType
	Id   string
}
