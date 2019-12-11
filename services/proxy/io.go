package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
	"net/http"
)

func (s *service) writeMembuffResponse(w http.ResponseWriter, message membuffers.Message, httpCode int, errorForVerbosity error) {
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

func (s *service) writeErrorResponseAndLog(w http.ResponseWriter, m *HttpErr) {
	if m.LogField == nil {
		s.logger.Info(m.Message)
	} else {
		s.logger.Info(m.Message, m.LogField)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(m.Code)
	_, err := w.Write([]byte(m.Message))
	if err != nil {
		s.logger.Info("error writing response", log.Error(err))
	}
}

func readInput(r *http.Request) ([]byte, *HttpErr) {
	if r.Body == nil {
		return nil, &HttpErr{http.StatusBadRequest, nil, "http request body is empty"}
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &HttpErr{http.StatusBadRequest, log.Error(err), "http request body is empty"}
	}
	return bytes, nil
}
