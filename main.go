package main

import (
	"http-server/internal"
	"log"
	"os"
)

const (
	Port              = ":8080"
	TransportProtocol = "tcp"
)

func main() {
	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags)
	s := internal.NewServer(logger)
	s.RegisterRoutes()
	s.Start(TransportProtocol, Port)
	s.Logger.Printf("HTTP Server has started and listening on port %s", Port)
}
