package main

import (
	"fmt"
	"io/ioutil"
	"net"
)

type NodeType string

const (
	NODE_VALIDATOR NodeType = "validator"
	NODE_NORMAL             = "normal"
)

type Node struct {
	ID   int
	Type NodeType
}

func (node *Node) GetAddress() string {
	return fmt.Sprintf("localhost:%d", node.ID)
}

func Start(id int) {
	node := Node{ID: id, Type: NODE_NORMAL}

	listener, err := net.Listen("tcp", node.GetAddress())
	Handle(err)

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		Handle(err)
		go HandleConnection(conn)
	}

}

func HandleConnection(conn net.Conn) {
	req, err := ioutil.ReadAll(conn)
	Handle(err)

	command := BytesToCmd(req[:commandLength])
	fmt.Println("Requested Command: " + command)
	switch command {
	default:
	}
}
