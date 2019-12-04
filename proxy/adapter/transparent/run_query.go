package transparent

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/proxy"
	"net/http"
)

func runQuery(h *handler, input []byte) (membuffers.Message, *proxy.HttpErr) {
	clientRequest := client.RunQueryRequestReader(input)
	if e := validate(clientRequest); e != nil {
		return nil, e
	}

	//h.logger.Info("http HttpServer received run-query", log.Stringable("request", clientRequest))

	res, resBody, err := sendHttpPost(h.config.Endpoints[0]+h.path, input)
	if err != nil {
		return nil, &proxy.HttpErr{http.StatusBadRequest, nil, err.Error()}
	}

	if res.StatusCode != http.StatusOK {
		return nil, &proxy.HttpErr{res.StatusCode, nil, res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return client.RunQueryResponseReader(resBody), nil
}
