package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"net/http"
)

func getTransactionStatus(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr) {
	input = client.GetTransactionStatusRequestReader(data)
	if e := validate(input); e != nil {
		return nil, nil, e
	}

	h.logger.Info("received request", log.Stringable("request", input))
	res, resBody, e := h.transport.SendRandom(h.config.Endpoints, h.path, data)
	if e != nil {
		return input, nil, &HttpErr{http.StatusBadRequest, log.Error(e), e.Error()}
	}

	if res.StatusCode != http.StatusOK {
		return input, nil, &HttpErr{res.StatusCode, log.Error(errors.New(res.Status)), res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return input, client.GetTransactionStatusResponseReader(resBody), &HttpErr{Code: res.StatusCode}
}
