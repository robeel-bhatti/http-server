package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PathParamKey string

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

const (
	Port                     = ":8080"
	Protocol                 = "HTTP/1.1"
	TmpDir                   = "tmp/"
	TextPlain                = "text/plain"
	OctetStream              = "application/octet-stream"
	PathParam   PathParamKey = "pathParams"
)

func main() {

	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags)
	s := &Server{Logger: logger}

	s.AddRoute("GET", "/", s.DefaultHandler)
	s.AddRoute("GET", "/echo/:name", s.EchoHandler)
	s.AddRoute("GET", "/files/:name", s.FilesHandler)
	s.AddRoute("POST", "/files/:name", s.FilesHandler)
	s.AddRoute("GET", "/user-agent", s.UserAgentHandler)

	// first create TCP listener
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		panic(err) // panicking here because code can't run without listener
	}
	defer listener.Close()

	s.Logger.Printf("Listening on %s", Port)

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
		s.WriteResponse(conn, s.BuildResponse(400, TextPlain, "could not read request"))
		return
	}

	handler := s.DetermineHandler(req)
	if handler == nil {
		s.Logger.Printf("No handler found for request with path %s\n", req.URL.Path)
		s.WriteResponse(conn, s.BuildResponse(404, TextPlain, "the request path is not supported"))
		return
	}

	s.WriteResponse(conn, handler(req))
	return
}

func (s *Server) BuildResponse(sc int, ct, body string) string {
	var sb strings.Builder
	sb.WriteString(Protocol + " " + strconv.Itoa(sc) + " " + http.StatusText(sc))
	sb.WriteString("\r\n")
	sb.WriteString("Content-Type: " + ct)
	sb.WriteString("\r\n")
	sb.WriteString("Content-Length: " + strconv.Itoa(len(body)))
	sb.WriteString("\r\n\r\n")
	sb.WriteString(body)
	return sb.String()
}

func (s *Server) WriteResponse(conn net.Conn, res string) {
	_, err := conn.Write([]byte(res))
	if err != nil {
		s.Logger.Printf("Error writing response: %v\n", err)
	}
}

// AddRoute registers a new URL on the server
func (s *Server) AddRoute(method, path string, handler Handler) {
	newRoute := &Route{
		handler: handler,
		method:  method,
		parts:   strings.Split(strings.Trim(path, "/"), "/"),
	}
	s.Routes = append(s.Routes, newRoute)
}

// DetermineHandler function determines the correct handler to handle the request
// this is done by validating that the request method matches the registered routes
// request method and matches every part of the registered routes URL (excluding placeholder values for path variables)
func (s *Server) DetermineHandler(req *http.Request) Handler {
	method := req.Method
	path := req.URL.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	routes := s.Routes

	for _, route := range routes {
		if method != route.method {
			continue
		}

		if len(pathParts) != len(route.parts) {
			continue
		}

		match := true
		pathParams := make(map[string]string)

		for i, routePart := range route.parts {
			if strings.HasPrefix(routePart, ":") {
				n := strings.TrimPrefix(routePart, ":")
				pathParams[n] = pathParts[i]
			} else if routePart != pathParts[i] {
				match = false
				break
			}
		}
		if match {
			ctx := context.WithValue(req.Context(), PathParam, pathParams)
			*req = *req.WithContext(ctx)
			return route.handler
		}
	}
	return nil
}

func (s *Server) DefaultHandler(r *http.Request) string {
	return s.BuildResponse(200, TextPlain, "")
}

// EchoHandler handles requests at /echo/{value}
func (s *Server) EchoHandler(r *http.Request) string {
	pv := r.Context().Value(PathParam).(map[string]string)["name"]
	return s.BuildResponse(200, TextPlain, pv)
}

func (s *Server) UserAgentHandler(r *http.Request) string {
	h := r.Header.Get("User-Agent")
	return s.BuildResponse(200, TextPlain, h)
}

func (s *Server) FilesHandler(r *http.Request) string {
	sc := 200
	val := r.Context().Value(PathParam).(map[string]string)
	f := TmpDir + val["name"]

	if r.Method == "POST" {
		err := os.WriteFile(f, []byte(val["name"]), 0644)
		if err != nil {
			s.Logger.Printf("error writing file at path %s: %v\n", f, err)
			return s.BuildResponse(500, TextPlain, "the requested file was not created")
		}
		sc = 201
	}

	data, err := os.ReadFile(f)
	if err != nil {
		s.Logger.Printf("error reading file at path %s: %v\n", f, err)
		return s.BuildResponse(404, TextPlain, "the requested file was not found")
	}

	return s.BuildResponse(sc, OctetStream, string(data))
}
