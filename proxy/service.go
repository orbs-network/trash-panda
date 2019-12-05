package proxy

import (
	"fmt"
	"github.com/orbs-network/scribe/log"
	"github.com/orbs-network/trash-panda/boostrap/httpserver"
	"net/http"
)

type Service struct {
	logger  log.Logger
	config  Config
	adapter ProxyAdapter
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
	}
}

func (s *Service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}

func (s *Service) UpdateRoutes(server *httpserver.HttpServer) {
	for _, h := range s.adapter.Handlers() {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h.Handler()))
	}
}

func (s *Service) wrapHandler(handlerBuilder HandlerBuilderFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, e := readInput(r)
		if e != nil {
			s.writeErrorResponseAndLog(w, e)
			return
		}
		result, err := handlerBuilder(bytes)
		if err != nil {
			s.writeErrorResponseAndLog(w, err)
		} else {
			s.writeMembuffResponse(w, result, 200, nil)
		}
	}
}
