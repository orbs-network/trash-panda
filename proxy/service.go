package proxy

import (
	"context"
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
	"github.com/orbs-network/trash-panda/config"
	"github.com/orbs-network/trash-panda/transport"
	"net/http"
	"time"
)

type Service struct {
	logger   log.Logger
	config   Config
	handlers []Handler

	queue chan membuffers.Message
}

type Config struct {
	VirtualChainId uint32
	Endpoints      []string
}

func NewService(cfg Config, transport transport.Transport, logger log.Logger) *Service {
	return &Service{
		config:   cfg,
		logger:   logger,
		handlers: GetHandlers(cfg, transport),
		queue:    make(chan membuffers.Message),
	}
}

func (s *Service) UpdateRoutes(server *httpserver.HttpServer) {
	for _, h := range s.handlers {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h.Handler()))
	}
}

func (s *Service) ResendTxQueue(ctx context.Context) {
	handle := govnr.Forever(ctx, "http server", config.NewErrorHandler(s.logger), func() {
		select {
		case message := <-s.queue:
			s.logger.Info("received callback", log.Stringable("message", message))
			switch message.(type) {
			case *client.SendTransactionRequest:
				input, _, err := s.findHandler("send-transaction").Handler()(message.Raw())
				if err != nil {
					go s.txCollectionCallback(input)
				}
			}

			<-time.After(100 * time.Millisecond)
		case <-ctx.Done():
			close(s.queue)
			s.logger.Info("shutting down")
			return
		}
	})

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(handle)
}

func (s *Service) txCollectionCallback(message membuffers.Message) {
	s.queue <- message
}

func (s *Service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}

func (s *Service) wrapHandler(handlerBuilder HandlerBuilderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, e := readInput(r)
		if e != nil {
			s.writeErrorResponseAndLog(w, e)
			return
		}
		input, output, err := handlerBuilder(bytes)

		//s.logger.Info("received request", log.Stringable("request", input))
		s.txCollectionCallback(input)

		if err != nil {
			s.logger.Error("error occurred", err.LogField)
			s.writeErrorResponseAndLog(w, err)
		} else {
			s.logger.Info("responded with", log.Stringable("response", output))
			s.writeMembuffResponse(w, output, 200, nil)
		}
	}
}

func (s *Service) findHandler(name string) Handler {
	for _, h := range s.handlers {
		if h.Name() == name {
			return h
		}
	}

	return nil
}
