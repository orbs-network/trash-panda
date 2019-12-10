package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/transport"
	"net/http"
)

type HandlerBuilderFunc func(data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr)

type Handler interface {
	Name() string
	Path() string
	Handle(data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr)
}

type builderFunc func(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr)

type handler struct {
	name      string
	path      string
	config    Config
	transport transport.Transport

	f builderFunc
}

func GetHandlers(config Config, transport transport.Transport) []Handler {
	// FIXME add more handlers
	//server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION_ASYNC), true, s.sendTransactionAsyncHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_RECEIPT_PROOF), true, s.getTransactionReceiptProofHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_BLOCK), true, s.getBlockHandler)

	return []Handler{
		&handler{
			name:      "run-query",
			path:      "/api/v1/run-query",
			config:    config,
			transport: transport,
			f:         runQuery,
		},
		&handler{
			name:      "send-transaction",
			path:      "/api/v1/send-transaction",
			config:    config,
			transport: transport,
			f:         sendTransaction,
		},
		&handler{
			name:      "get-transaction-status",
			path:      "/api/v1/get-transaction-status",
			config:    config,
			transport: transport,
			f:         getTransactionStatus,
		},
	}
}

func (h *handler) Name() string {
	return h.name
}

func (h *handler) Path() string {
	return h.path
}

func (h *handler) Handle(data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr) {
	return h.f(h, data)
}

func validate(m membuffers.Message) *HttpErr {
	if !m.IsValid() {
		return &HttpErr{http.StatusBadRequest, log.Stringable("request", m), "http request is not a valid membuffer"}
	}
	return nil
}
