package main

import (
	"http-server/internal"
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags)
	s := internal.NewServer(logger)
	s.RegisterRoutes()
	s.Start()
}
