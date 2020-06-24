package model

type NodeType string

const (
	Server  NodeType = "server"
	Device  NodeType = "device"
	Cli     NodeType = "cli"
	Browser NodeType = "browser"
)

type Node struct {
	Node NodeType
	Id   string
}
