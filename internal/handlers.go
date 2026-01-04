package internal

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

var logger *log.Logger
var compressionSchemes = []string{"gzip", "deflate"}

func SetLogger(l *log.Logger) {
	logger = l
}

func DefaultHandler(r *http.Request) *ResponseEntity {
	logger.Println("DefaultHandler called")
	return NewResponseEntity(http.StatusOK, TextPlain, "", "")
}

func GetUserAgentHandler(r *http.Request) *ResponseEntity {
	logger.Println("GetUserAgentHandler called")
	return NewResponseEntity(http.StatusOK, TextPlain, "", r.Header.Get("User-Agent"))
}

func GetEchoStringHandler(r *http.Request) *ResponseEntity {
	logger.Println("GetEchoStringHandler called")
	pv := getPathParam(r, "name")
	ce := selectEncoding(r.Header.Get("Accept-Encoding"))
	b, err := compressBody(pv, ce)
	if err != nil {
		logger.Printf("error compressing Body: %v", err)
		ce = ""
		b = "unexpected error occurred"
	}
	return NewResponseEntity(http.StatusOK, TextPlain, ce, b)
}

func ReadFileHandler(r *http.Request) *ResponseEntity {
	logger.Println("ReadFileHandler called")

	pv := getPathParam(r, "pv")
	if !validFileName(pv) {
		logger.Printf("invalid file pv in request: %s", pv)
		return NewResponseEntity(http.StatusBadRequest, TextPlain, "", "invalid file pv provided")
	}

	fp := TmpDir + pv
	d, err := os.ReadFile(fp)
	if err != nil {
		logger.Printf("error reading file %s: %v", fp, err)
		return NewResponseEntity(http.StatusNotFound, TextPlain, "", "file not found")
	}
	return NewResponseEntity(http.StatusOK, OctetStream, "", string(d))
}

func WriteFileHandler(r *http.Request) *ResponseEntity {
	logger.Println("WriteFileHandler called")

	pv := getPathParam(r, "pv")
	if !validFileName(pv) {
		logger.Printf("invalid file pv in request: %s", pv)
		return NewResponseEntity(http.StatusBadRequest, TextPlain, "", "invalid file pv provided")
	}

	fp := TmpDir + pv
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Printf("error occurred reading request Body: %v", err)
		return NewResponseEntity(http.StatusInternalServerError, TextPlain, "", "unexpected error occurred")
	}

	err = os.WriteFile(fp, b, 0644)
	if err != nil {
		logger.Printf("error occurred writing file at path %s: %v", fp, err)
		return NewResponseEntity(http.StatusInternalServerError, TextPlain, "", "unexpected error occurred")
	}
	return NewResponseEntity(http.StatusCreated, OctetStream, "", "")
}

func selectEncoding(ae string) string {
	if ae == "" {
		return ""
	}
	schemes := strings.Split(ae, ",")
	for _, scheme := range schemes {
		scheme = strings.TrimSpace(scheme)
		if slices.Contains(compressionSchemes, scheme) {
			return scheme
		}
	}
	return ""
}

func compressBody(b, ce string) (string, error) {
	var buffer bytes.Buffer

	switch ce {
	case "gzip":
		gw := gzip.NewWriter(&buffer)
		if _, err := gw.Write([]byte(b)); err != nil {
			return "", err
		}
		if err := gw.Close(); err != nil {
			return "", err
		}
		return buffer.String(), nil
	case "deflate":
		fw, _ := flate.NewWriter(&buffer, -1)
		if _, err := fw.Write([]byte(b)); err != nil {
			return "", err
		}
		if err := fw.Close(); err != nil {
			return "", err
		}
		return buffer.String(), nil
	default:
		return b, nil
	}
}

func validFileName(name string) bool {
	if name == "" || strings.Contains(name, "..") { // ".." check to prevent path traversals
		return false
	}
	return true
}

func getPathParam(r *http.Request, key string) string {
	pathParamMap := r.Context().Value(CustomPathParamKey).(map[string]string)
	return pathParamMap[key]
}
