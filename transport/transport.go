package transport

import "net/http"

type Transport interface {
	Send(endpoint string, path string, payload []byte) (*http.Response, []byte, error)
	SendRandom(endpoints []string, path string, payload []byte) (*http.Response, []byte, error)
}
