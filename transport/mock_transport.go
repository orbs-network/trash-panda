package transport

import (
	"github.com/pkg/errors"
	"net/http"
)

type MockTransport struct {
	on            bool
	httpTransport Transport
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		on:            true,
		httpTransport: NewHttpTransport(),
	}
}

func (t *MockTransport) Send(endpoint string, payload []byte) (*http.Response, []byte, error) {
	if t.on {
		return t.httpTransport.Send(endpoint, payload)
	}

	return nil, nil, errors.New("failed to send payload")
}

func (t *MockTransport) On() {
	t.on = true
}

func (t *MockTransport) Off() {
	t.on = false
}
