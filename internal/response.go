package internal

import (
	"net/http"
	"strconv"
	"strings"
)

type ResponseEntity struct {
	statusCode int
	headers    *HttpHeader
	body       string
}

type HttpHeader struct {
	ContentType     string
	ContentEncoding string
	ContentLength   string
}

func NewResponseEntity(sc int, ct, ce, body string) *ResponseEntity {
	return &ResponseEntity{
		statusCode: sc,
		headers:    NewHttpHeader(ct, ce, len(body)),
		body:       body,
	}
}

func NewHttpHeader(ct, ce string, cl int) *HttpHeader {
	return &HttpHeader{
		ContentType:     ct,
		ContentEncoding: ce,
		ContentLength:   strconv.Itoa(cl),
	}
}

func HttpResponseBuilder(statusCode int, contentType, contentEncoding, body string) string {
	var sb strings.Builder

	sb.WriteString(HttpProtocol + " " + strconv.Itoa(statusCode) + " " + http.StatusText(statusCode))
	sb.WriteString("\r\n")
	sb.WriteString("Content-Type: " + contentType)
	sb.WriteString("\r\n")

	if contentEncoding != "" {
		body = compressHttpBody(contentEncoding, body)
		sb.WriteString("Content-Encoding: " + contentEncoding)
		sb.WriteString("\r\n")
	}

	sb.WriteString("Content-Length: " + strconv.Itoa(len(body)))
	sb.WriteString("\r\n\r\n")
	sb.WriteString(body)
	return sb.String()
}
