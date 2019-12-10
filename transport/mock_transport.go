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

func (t *MockTransport) Send(endpoint string, payload []byte) (*http.Response, []byte, error) {
	if t.on {
		println("sending payload to", endpoint)
		return t.httpTransport.Send(endpoint, payload)
	}

	println("skipping payload to", endpoint)
	return nil, nil, errors.New("failed to send payload")
}

func (t *MockTransport) On() {
	t.on = true
}

func (t *MockTransport) Off() {
	t.on = false
}
