package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"net/http"
)

type HttpErr struct {
	Code     int
	LogField *log.Field
	Message  string
}

func (e *HttpErr) ToError() error {
	if e.Message != "" {
		return errors.New(e.Message)
	}

	return nil
}

func validate(m membuffers.Message) *HttpErr {
	if !m.IsValid() {
		return &HttpErr{http.StatusBadRequest, log.Stringable("request", m), "http request is not a valid membuffer"}
	}
	return nil
}
