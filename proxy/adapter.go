package proxy

import "github.com/orbs-network/membuffers/go"

type HandlerBuilderFunc func(input []byte) (message membuffers.Message, err *HttpErr)

type Handler interface {
	Name() string
	Path() string
	Handler() HandlerBuilderFunc
}

type ProxyAdapter interface {
	Handlers() []Handler
}
