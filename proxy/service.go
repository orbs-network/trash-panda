package proxy

import (
	"context"
	"fmt"
	"github.com/orbs-network/membuffers/go"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
	"net/http"
)

type Service struct {
	logger  log.Logger
	config  Config
	adapter ProxyAdapter

	queue chan membuffers.Message
}

type Config struct {
	VirtualChainId uint32
	Endpoints      []string
}

func NewService(cfg Config, adapter ProxyAdapter, logger log.Logger) *Service {
	return &Service{
		config:  cfg,
		logger:  logger,
		adapter: adapter,
		queue:   make(chan membuffers.Message),
	}
}

func (s *Service) UpdateRoutes(server *httpserver.HttpServer) {
	for _, h := range s.adapter.Handlers() {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h.Handler()))
	}
}

func (s *Service) ResendTxQueue(ctx context.Context) {
	for {
		select {
		case message := <-s.queue:
			s.logger.Info("received callback", log.Stringable("message", message))
		case <-ctx.Done():
			close(s.queue)
			s.logger.Info("shutting down")
			return
		}
	}
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

		s.logger.Info("received request", log.Stringable("request", input))
		if err != nil {
			s.logger.Error("error occurred", err.LogField)
			s.writeErrorResponseAndLog(w, err)
		} else {
			s.logger.Info("responded with", log.Stringable("response", output))
			s.writeMembuffResponse(w, output, 200, nil)
		}
	}
}
