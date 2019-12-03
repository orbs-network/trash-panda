package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
	"net/http"
)

type httpErr struct {
	code     int
	logField *log.Field
	message  string
}

func (s *Service) writeMembuffResponse(w http.ResponseWriter, message membuffers.Message, httpCode int, errorForVerbosity error) {
	w.Header().Set("Content-Type", "application/membuffers")

	if errorForVerbosity != nil {
		w.Header().Set("X-ORBS-ERROR-DETAILS", errorForVerbosity.Error())
	}
	w.WriteHeader(httpCode)
	_, err := w.Write(message.Raw())
	if err != nil {
		s.logger.Info("error writing response", log.Error(err))
	}
}

func (s *Service) writeErrorResponseAndLog(w http.ResponseWriter, m *httpErr) {
	if m.logField == nil {
		s.logger.Info(m.message)
	} else {
		s.logger.Info(m.message, m.logField)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(m.code)
	_, err := w.Write([]byte(m.message))
	if err != nil {
		s.logger.Info("error writing response", log.Error(err))
	}
}

func readInput(r *http.Request) ([]byte, *httpErr) {
	if r.Body == nil {
		return nil, &httpErr{http.StatusBadRequest, nil, "http request body is empty"}
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &httpErr{http.StatusBadRequest, log.Error(err), "http request body is empty"}
	}
	return bytes, nil
}

func validate(m membuffers.Message) *httpErr {
	if !m.IsValid() {
		return &httpErr{http.StatusBadRequest, log.Stringable("request", m), "http request is not a valid membuffer"}
	}
	return nil
}

var HEADERS = []string{
	"Content-Type",
	"X-ORBS-REQUEST-RESULT",
	"X-ORBS-BLOCK-HEIGHT",
	"X-ORBS-BLOCK-TIMESTAMP",
	"X-ORBS-ERROR-DETAILS",
}

func copyResponse(w http.ResponseWriter, res *http.Response, responseBody []byte) {
	w.WriteHeader(res.StatusCode)
	w.Write(responseBody)

	for _, header := range HEADERS {
		println(header, res.Header.Get(header))
		w.Header().Set(header, res.Header.Get(header))
	}
	w.Header().Write(w)

	println("====")
	for _, header := range HEADERS {
		println(header, w.Header().Get(header))
	}
}
