package transparent

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/trash-panda/proxy"
	"net/http"
)

func runQuery(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *proxy.HttpErr) {
	input = client.RunQueryRequestReader(data)
	if e := validate(input); e != nil {
		return nil, nil, e
	}

	res, resBody, e := sendHttpPost(h.config.Endpoints[0]+h.path, data)
	if e != nil {
		return input, nil, &proxy.HttpErr{http.StatusBadRequest, nil, e.Error()}
	}

	if res.StatusCode != http.StatusOK {
		return input, nil, &proxy.HttpErr{res.StatusCode, nil, res.Header.Get("X-ORBS-ERROR-DETAILS")}
	}

	return input, client.RunQueryResponseReader(resBody), nil
}
