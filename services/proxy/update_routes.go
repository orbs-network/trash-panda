package proxy

import (
	"fmt"
	"github.com/orbs-network/trash-panda/bootstrap/httpserver"
)

func (s *service) UpdateRoutes(server *httpserver.HttpServer) {
	for _, h := range s.handlers {
		server.RegisterHttpHandler(server.Router(), s.getPath(h.Path()), true, s.wrapHandler(h))
	}
}

func (s *service) getPath(path string) string {
	return fmt.Sprintf("/vchains/%d%s", s.config.VirtualChainId, path)
}
