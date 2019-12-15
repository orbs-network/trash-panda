package proxy

import (
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/transport"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

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
	logger    log.Logger

	f builderFunc
}

func GetHandlers(config Config, transport transport.Transport, logger log.Logger) []Handler {
	// FIXME add more handlers
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_TRANSACTION_RECEIPT_PROOF), true, s.getTransactionReceiptProofHandler)
	//server.RegisterHttpHandler(server.Router(), s.getPath(GET_BLOCK), true, s.getBlockHandler)

	return []Handler{
		&handler{
			name:      "run-query",
			path:      "/api/v1/run-query",
			config:    config,
			transport: transport,
			logger:    logger,
			f:         runQuery,
		},
		&handler{
			name:      "send-transaction",
			path:      "/api/v1/send-transaction",
			config:    config,
			transport: transport,
			logger:    logger,
			f:         sendTransaction,
		},
		&handler{
			name:      "send-transaction-async",
			path:      "/api/v1/send-transaction-async",
			config:    config,
			transport: transport,
			logger:    logger,
			f:         sendTransactionAsync,
		},
		&handler{
			name:      "get-transaction-status",
			path:      "/api/v1/get-transaction-status",
			config:    config,
			transport: transport,
			logger:    logger,
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

func (h *handler) getShuffledEndpoints() []string {
	shuffled := make([]string, len(h.config.Endpoints))
	copy(shuffled, h.config.Endpoints)

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func (s *service) wrapHandler(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, e := readInput(r)
		if e != nil {
			s.writeErrorResponseAndLog(w, e)
			return
		}
		input, output, err := h.Handle(bytes)

		if err := s.storeInput(input, output); err != nil {
			s.logger.Error("failed to store incoming transaction", log.Error(err))
		}

		if output != nil {
			s.logger.Info("responded with", log.Stringable("response", output))
			s.writeMembuffResponse(w, output, err.Code, nil)
		} else {
			s.logger.Error("error occurred", err.LogField)
			s.writeErrorResponseAndLog(w, err)
		}
	}
}

func (s *service) storeInput(input membuffers.Message, output membuffers.Message) error {
	switch input.(type) {
	case *client.SendTransactionRequest:
		signedTransaction := input.(*client.SendTransactionRequest).SignedTransaction()
		transactionStatus := protocol.TRANSACTION_STATUS_RESERVED

		if res, ok := output.(*client.SendTransactionResponse); ok {
			transactionStatus = res.TransactionStatus()
		}

		return s.storage.StoreTransaction(signedTransaction, transactionStatus)
	}

	return nil
}

func (s *service) findHandler(name string) Handler {
	for _, h := range s.handlers {
		if h.Name() == name {
			return h
		}
	}

	return nil
}

func aggregateRequest(times int, endpoints []string, request func(endpoint string) response) []response {
	var results []response

	maxTimes := times
	if maxTimes > len(endpoints) {
		maxTimes = len(endpoints)
	}
	for i := 0; i < maxTimes; i++ {
		results = append(results, request(endpoints[i]))
	}

	return results
}

func filterResponses(responses []response) response {
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].blockHeight > responses[j].blockHeight
	})

	return responses[0]
}

type response struct {
	output      membuffers.Message
	httpErr     *HttpErr
	blockHeight uint64
}
