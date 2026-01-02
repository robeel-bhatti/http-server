package internal

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
)

type HttpHeader struct {
	ContentType     string
	ContentEncoding string
	ContentLength   string
}

type ResponseEntity struct {
	statusCode int
	headers    *HttpHeader
	body       string
	err        error
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

func compressHttpBody(contentEncoding, body string) string {
	if contentEncoding == "gzip" {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		w.Write([]byte(body))
		w.Close()
		return buf.String()
	}
	return body
}

func getPathParam(r *http.Request, key string) string {
	pathParamMap := r.Context().Value(CustomPathParamKey).(map[string]string)
	return pathParamMap[key]
}
