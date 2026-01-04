package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type PathParamContextKey string

type Handler func(r *http.Request) *ResponseEntity

type Route struct {
	handler Handler
	method  string
	parts   []string // the parts of the route ["user", ":id", "posts"]
}

type Server struct {
	Logger *log.Logger
	Routes []*Route
}

type ResponseEntity struct {
	StatusCode int
	Headers    *HttpHeader
	Body       string
}

type HttpHeader struct {
	ContentType     string
	ContentEncoding string
	ContentLength   string
}

func NewServer(logger *log.Logger) *Server {
	return &Server{Logger: logger}
}

func (s *Server) Start() {
	listener, err := net.Listen(TransportProtocol, Port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	s.Logger.Printf("HTTP Server has started and listening on port %s", Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.Logger.Printf("error accepting request: %v\n", err)
			continue // this time just log and continue, don't stop server here
		}
		go s.HandleConnection(conn)
	}
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close() // close the TCP connection once the response has been delivered to the client

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		s.Logger.Printf("error reading request: %v", err)
		re := NewResponseEntity(http.StatusInternalServerError, TextPlain, "", "unexpected error occurred")
		s.BuildAndSendResponse(conn, re)
		return
	}

	handler := s.DetermineHandler(req)
	if handler == nil {
		msg := fmt.Sprintf("the requested resource %s is not supported", req.URL.Path)
		s.Logger.Printf(msg)
		re := NewResponseEntity(http.StatusNotFound, TextPlain, "", msg)
		s.BuildAndSendResponse(conn, re)
		return
	}

	SetLogger(s.Logger)
	re := handler(req)
	s.BuildAndSendResponse(conn, re)
	return
}

func (s *Server) RegisterRoutes() {
	s.AddRoute("GET", "/", DefaultHandler)
	s.AddRoute("GET", "/echo/:name", GetEchoStringHandler)
	s.AddRoute("GET", "/user-agent", GetUserAgentHandler)
	s.AddRoute("GET", "/files/:name", ReadFileHandler)
	s.AddRoute("POST", "/files/:name", WriteFileHandler)
}

func (s *Server) AddRoute(method, pattern string, handler Handler) {
	newRoute := &Route{
		handler: handler,
		method:  method,
		parts:   strings.Split(strings.Trim(pattern, "/"), "/"),
	}
	s.Routes = append(s.Routes, newRoute)
}

func (s *Server) DetermineHandler(req *http.Request) Handler {
	method := req.Method
	path := req.URL.Path
	routes := s.Routes
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

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
			ctx := context.WithValue(req.Context(), CustomPathParamKey, pathParams)
			*req = *req.WithContext(ctx)
			return route.handler
		}
	}
	return nil
}

func (s *Server) BuildAndSendResponse(conn net.Conn, re *ResponseEntity) {
	var sb strings.Builder

	sb.WriteString(HttpProtocol + " " + strconv.Itoa(re.StatusCode) + " " + http.StatusText(re.StatusCode))
	sb.WriteString("\r\n")
	sb.WriteString("Content-Type: " + re.Headers.ContentType)
	sb.WriteString("\r\n")
	sb.WriteString("Content-Length: " + re.Headers.ContentLength)
	sb.WriteString("\r\n")

	if re.Headers.ContentEncoding != "" {
		sb.WriteString("Content-Encoding: " + re.Headers.ContentEncoding)
		sb.WriteString("\r\n")
	}

	sb.WriteString("\r\n\r\n")
	if re.Body != "" {
		sb.WriteString(re.Body)
	}

	if _, err := conn.Write([]byte(sb.String())); err != nil {
		s.Logger.Printf("error writing response: %v\n", err)
	}
}

func NewResponseEntity(sc int, ct, ce, body string) *ResponseEntity {
	return &ResponseEntity{
		StatusCode: sc,
		Headers:    NewHttpHeader(ct, ce, len(body)),
		Body:       body,
	}
}

func NewHttpHeader(ct, ce string, cl int) *HttpHeader {
	return &HttpHeader{
		ContentType:     ct,
		ContentEncoding: ce,
		ContentLength:   strconv.Itoa(cl),
	}
}
