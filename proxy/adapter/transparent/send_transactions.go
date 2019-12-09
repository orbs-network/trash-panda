package transparent

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/proxy"
	"github.com/pkg/errors"
	"net/http"
)

func sendTransaction(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *proxy.HttpErr) {
	input = client.SendTransactionRequestReader(data)
	if e := validate(input); e != nil {
		return nil, nil, e
	}

	res, resBody, e := h.transport.Send(h.config.Endpoints[0]+h.path, data)
	if e != nil {
		return input, nil, &proxy.HttpErr{http.StatusBadRequest, log.Error(e), e.Error()}
	}

	if res.StatusCode != http.StatusOK {
		//readerResponse := (&client.SendTransactionResponseBuilder{
		//	TransactionStatus: protocol.TRANSACTION_STATUS_PENDING,
		//	RequestResult: &client.RequestResultBuilder{
		//		RequestStatus: protocol.REQUEST_STATUS_IN_PROCESS,
		//	},
		//}).Build()
		return input, nil, &proxy.HttpErr{res.StatusCode, log.Error(errors.New(res.Status)), res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return input, client.SendTransactionResponseReader(resBody), nil
}
