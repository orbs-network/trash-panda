package transparent

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/proxy"
	"net/http"
)

type adapter struct {
	handlers []proxy.Handler
}

type builderFunc func(h *handler, data []byte) (input membuffers.Message, output membuffers.Message, err *proxy.HttpErr)

type handler struct {
	name   string
	path   string
	config proxy.Config

	f builderFunc
}

func NewTransparentAdapter(config proxy.Config) proxy.ProxyAdapter {
	// FIXME add more handlers
	//server.RegisterHttpHandler(server.Router(), s.getPath(SEND_TRANSACTION_ASYNC), true, s.sendTransactionAsyncHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_STATUS), true, s.getTransactionStatusHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_RECEIPT_PROOF), true, s.getTransactionReceiptProofHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_BLOCK), true, s.getBlockHandler)

	return &adapter{
		[]proxy.Handler{
			&handler{
				name:   "run-query",
				path:   "/api/v1/run-query",
				config: config,
				f:      runQuery,
			},
			&handler{
				name:   "send-transaction",
				path:   "/api/v1/send-transaction",
				config: config,
				f:      sendTransaction,
			},
		},
	}
}

func (a *adapter) Handlers() []proxy.Handler {
	return a.handlers
}

func (h *handler) Name() string {
	return h.name
}

func (h *handler) Path() string {
	return h.path
}

func (h *handler) Handler() proxy.HandlerBuilderFunc {
	return func(data []byte) (input membuffers.Message, output membuffers.Message, err *proxy.HttpErr) {
		return h.f(h, data)
	}
}

func validate(m membuffers.Message) *proxy.HttpErr {
	if !m.IsValid() {
		return &proxy.HttpErr{http.StatusBadRequest, log.Stringable("request", m), "http request is not a valid membuffer"}
	}
	return nil
}
