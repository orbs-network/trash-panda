package proxy

import "github.com/orbs-network/membuffers/go"

type HandlerBuilderFunc func(input []byte) (message membuffers.Message, err *HttpErr)
type HandlerCallback func(message membuffers.Message)

type Handler interface {
	Name() string
	Path() string
	Handler() HandlerBuilderFunc
	SetCallback(callback HandlerCallback)
}

type ProxyAdapter interface {
	Handlers() []Handler
}
