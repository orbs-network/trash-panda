package transport

import (
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type MockTransport struct {
	on            bool
	httpTransport Transport
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		on: true,
		httpTransport: NewHttpTransport(Config{
			Timeout: 1 * time.Second,
		}),
	}
}

func (t *MockTransport) Send(endpoint string, path string, payload []byte) (*http.Response, []byte, error) {
	if t.on {
		println("sending payload to", endpoint)
		return t.httpTransport.Send(endpoint, path, payload)
	}

	println("skipping payload to", endpoint)
	return nil, nil, errors.New("failed to send payload")
}

func (t *MockTransport) SendRandom(endpoints []string, path string, payload []byte) (*http.Response, []byte, error) {
	return t.Send(endpoints[0], path, payload)
}

func (t *MockTransport) On() {
	t.on = true
}

func (t *MockTransport) Off() {
	t.on = false
}
