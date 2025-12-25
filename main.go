package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	PORT = ":8080"
)

type Handler func(r *http.Request) string

type Route struct {
	handler Handler
	method  string
	parts   []string // the parts of the route ["user", ":id", "posts"]
}

type Server struct {
	Logger *log.Logger
	Routes []*Route
}

func (s *Server) AddRoute(method, path string, handler Handler) {
	newRoute := &Route{
		handler: handler,
		method:  method,
		parts:   strings.Split(strings.Trim(path, "/"), "/"),
	}

	s.Routes = append(s.Routes, newRoute)
}

func (s *Server) DetermineHandler(method, path string) Handler {
	// first determine which routes match the provided method
	// we will filter down to those routes

	var handler Handler
	routes := s.Routes

	for _, route := range routes {
		if method == route.method {

			// determine which parts of the route match up with the provided path.
			// first thing we want to do is check if the slices are of equal length
			// this avoids index errors in the future checks
			pathParts := strings.Split(strings.Trim(path, "/"), "/")

			if len(pathParts) == len(route.parts) {

				// if slices are equal, check the values at each index for both slices is the same
				// IF the value does not start with ":" in the route.parts slice
				match := true
				for i, e := range pathParts {
					if !strings.HasPrefix(route.parts[i], ":") && e != route.parts[i] {
						match = false
						break
					}
				}
				if match {
					handler = route.handler
					break
				}

			}
		}
	}
	return handler
}

func main() {

	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags)
	s := &Server{Logger: logger}

	s.AddRoute("GET", "/", s.DefaultHandler)
	s.AddRoute("GET", "/echo/:name", s.EchoHandler)
	s.AddRoute("GET", "/files/:name", s.FilesHandler)
	s.AddRoute("GET", "/user-agent", s.UserAgentHandler)

	// first create TCP listener
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		panic(err) // panicking here because code can't run without listener
	}
	defer listener.Close()

	s.Logger.Printf("Listening on %s", PORT)

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
		_, err = conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		if err != nil {
			s.Logger.Printf("Error writing response: %v\n", err)
		}
		return
	}

	handler := s.DetermineHandler(req.Method, req.URL.Path)
	if handler == nil {
		s.Logger.Printf("No handler found for request with path %s\n", req.URL.Path)
		_, err = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		if err != nil {
			s.Logger.Printf("Error writing response: %v\n", err)
		}
		return
	}

	res := handler(req)
	_, err = conn.Write([]byte(res))
	if err != nil {
		s.Logger.Printf("Error writing response: %v\n", err)
	}
}

func (s *Server) DefaultHandler(r *http.Request) string {
	return "HTTP/1.1 200 OK\r\n\r\n"
}

// EchoHandler handles requests at /echo/{value}
func (s *Server) EchoHandler(r *http.Request) string {
	pv := strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1]

	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: " + strconv.Itoa(len(pv)) + "\r\n" +
		"\r\n" +
		pv
	return resp
}

func (s *Server) UserAgentHandler(r *http.Request) string {
	h := r.Header.Get("User-Agent")
	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: " + strconv.Itoa(len(h)) + "\r\n" +
		"\r\n" +
		h
	return resp
}

func (s *Server) FilesHandler(r *http.Request) string {
	f := fmt.Sprintf("tmp/%s", strings.Split(strings.Trim(r.URL.Path, "/"), "/")[1])
	data, err := os.ReadFile(f)
	if err != nil {
		s.Logger.Printf("error reading file: %v\n", err)
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: " + strconv.Itoa(len(string(data))) + "\r\n" +
		"\r\n" +
		string(data)

	return resp
}
