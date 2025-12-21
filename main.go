package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"os"
)

type Server struct {
	Logger *log.Logger
}

func main() {

	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags)
	s := &Server{Logger: logger}

	// first create TCP listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err) // panicking here because code can't run without listener
	}
	defer listener.Close()

	s.Logger.Println("Listening on :8080")

	for {
		c, err := listener.Accept()
		if err != nil {
			s.Logger.Printf("Error accepting request: %v\n", err)
			continue // this time just log and continue, don't stop server here
		}

		go s.HandleConnection(c)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		s.Logger.Printf("Error reading request: %v\n", err)
		return
	}

	switch req.URL.Path {
	case "/":
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		if err != nil {
			s.Logger.Printf("Error writing response for '/' path: %v\n", err)
			return
		}
	default:
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		if err != nil {
			s.Logger.Printf("Error writing default response: %v\n", err)
			return
		}
	}
}
