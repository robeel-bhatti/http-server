package internal

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	TextPlain    = "text/plain"
	OctetStream  = "application/octet-stream"
	HttpProtocol = "HTTP/1.1"
	TmpDir       = "tmp/"
)

func DefaultHandler(r *http.Request) (string, error) {
	return httpResponseBuilder(200, TextPlain, "", ""), nil
}

func UserAgentHandler(r *http.Request) (string, error) {
	headerValue := r.Header.Get("User-Agent")
	return httpResponseBuilder(200, TextPlain, "", headerValue), nil
}

func EchoHandler(r *http.Request) (string, error) {
	pv := r.Context().Value(PathParam).(map[string]string)["name"]

	var ce string
	if r.Header.Get("Accept-Encoding") != "" {
		ae := r.Header.Get("Accept-Encoding")
		aeList := strings.Split(ae, ", ")
		for _, scheme := range aeList {
			if slices.Contains(getCompressionSchemes(), scheme) {
				ce = scheme
				break
			}
		}
	}
	return httpResponseBuilder(200, TextPlain, ce, pv), nil
}

func FilesHandler(r *http.Request) (string, error) {
	code := 200
	pathVar := r.Context().Value(PathParam).(map[string]string)["name"]
	filePath := TmpDir + pathVar

	if r.Method == "POST" {
		err := os.WriteFile(filePath, []byte(pathVar), 0644)
		if err != nil {
			return "", err
		}
		code = 201
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil
	}

	return httpResponseBuilder(code, OctetStream, "", string(data)), nil
}

func httpResponseBuilder(statusCode int, contentType, contentEncoding, body string) string {
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

func getCompressionSchemes() []string {
	return []string{"gzip"}
}
