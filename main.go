package main

import (
	"net"
)

func main() {
	// first create TCP listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	// then accept incoming TCP connections in a loop
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go HandleConnection(c)
	}
}

func HandleConnection(c net.Conn) {

}
