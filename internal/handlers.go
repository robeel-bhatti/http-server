package internal

import (
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

const (
	TextPlain    = "text/plain"
	OctetStream  = "application/octet-stream"
	HttpProtocol = "HTTP/1.1"
	TmpDir       = "tmp/"
)

func DefaultHandler(r *http.Request) *ResponseEntity {
	return &ResponseEntity{
		statusCode: http.StatusOK,
		headers:    NewHttpHeader(TextPlain, "", len("")),
		body:       "",
		err:        nil,
	}
}

func UserAgentHandler(r *http.Request) *ResponseEntity {
	headerValue := r.Header.Get("User-Agent")
	return &ResponseEntity{
		statusCode: http.StatusOK,
		headers:    NewHttpHeader(TextPlain, "", len(headerValue)),
		body:       headerValue,
		err:        nil,
	}
}

func EchoHandler(r *http.Request) *ResponseEntity {
	pv := getPathParam(r, "name")
	var ce string

	if r.Header.Get("Accept-Encoding") != "" {
		ae := r.Header.Get("Accept-Encoding")
		aeList := strings.Split(ae, ", ")
		validSchemes := []string{"gzip"}

		for _, scheme := range aeList {
			if slices.Contains(validSchemes, scheme) {
				ce = scheme
				break
			}
		}
	}

	return &ResponseEntity{
		statusCode: http.StatusOK,
		headers:    NewHttpHeader(TextPlain, ce, len(pv)),
		body:       pv,
		err:        nil,
	}
}

func FilesHandler(r *http.Request) *ResponseEntity {
	fp := TmpDir + getPathParam(r, "name")

	if r.Method == "GET" {
		data, err := os.ReadFile(fp)
		d := string(data)
		return &ResponseEntity{
			statusCode: 200,
			headers:    NewHttpHeader(OctetStream, "", len(d)),
			body:       d,
			err:        err,
		}
	} else {
		reqBytes, err := io.ReadAll(r.Body)
		err = os.WriteFile(fp, reqBytes, 0644)
		return &ResponseEntity{
			statusCode: 201,
			headers:    NewHttpHeader(OctetStream, "", len("")),
			body:       "",
			err:        err,
		}
	}
}
