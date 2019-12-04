package transparent

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/proxy"
	"net/http"
)

func sendTransaction(h *handler, input []byte) (membuffers.Message, *proxy.HttpErr) {
	clientRequest := client.SendTransactionRequestReader(input)
	if e := validate(clientRequest); e != nil {
		return nil, e
	}

	res, resBody, err := sendHttpPost(h.config.Endpoints[0]+h.path, input)
	if err != nil {
		return nil, &proxy.HttpErr{http.StatusBadRequest, nil, err.Error()}
	}

	if res.StatusCode != http.StatusOK {
		return nil, &proxy.HttpErr{res.StatusCode, nil, res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return client.SendTransactionResponseReader(resBody), nil
}
