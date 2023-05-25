package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

func NewHttpClientTimeout(timeout time.Duration) http.Client {
	return NewHttpClient(timeout, false)
}

func NewHttpClient(timeout time.Duration, insecureSkipVerify bool) http.Client {
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	client := http.Client{
		Transport: &transport,
		Timeout:   timeout,
	}

	return client
}

func WriteJsonResponse(w http.ResponseWriter, obj interface{}) {
	js, _ := json.Marshal(obj)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func WriteHttpError(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%d error : %v", code, error)
}

func WriteJsonStringResponse(w http.ResponseWriter, statusCode int, str string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(str))
}

func WriteBytes(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}
