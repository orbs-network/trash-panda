package proxy

import "github.com/orbs-network/membuffers/go"

type HandlerBuilderFunc func(data []byte) (input membuffers.Message, output membuffers.Message, err *HttpErr)

type Handler interface {
	Name() string
	Path() string
	Handler() HandlerBuilderFunc
}

type ProxyAdapter interface {
	Handlers() []Handler
}
