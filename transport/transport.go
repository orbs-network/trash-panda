package transport

import "net/http"

type Transport interface {
	Send(endpoint string, payload []byte) (*http.Response, []byte, error)
}
