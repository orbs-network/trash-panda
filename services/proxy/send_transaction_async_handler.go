package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"net/http"
)

func sendTransactionAsync(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr) {
	input = client.SendTransactionRequestReader(data)
	if e := validate(input); e != nil {
		return nil, nil, e
	}

	h.logger.Info("received request", log.Stringable("request", input))

	shuffledEndpoints := h.getShuffledEndpoints()
	res, resBody, e := h.transport.Send(shuffledEndpoints[0], h.path, data)
	if e != nil {
		output = (&client.SendTransactionResponseBuilder{
			TransactionStatus: protocol.TRANSACTION_STATUS_PENDING,
			RequestResult: &client.RequestResultBuilder{
				RequestStatus: protocol.REQUEST_STATUS_IN_PROCESS,
			},
		}).Build()

		return input, output, &HttpErr{http.StatusAccepted, log.Error(e), e.Error()}
	}

	if res.StatusCode != http.StatusAccepted {
		return input, nil, &HttpErr{res.StatusCode, log.Error(errors.New(res.Status)), res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return input, client.SendTransactionResponseReader(resBody), &HttpErr{Code: res.StatusCode}
}
