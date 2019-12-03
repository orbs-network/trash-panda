package httpserver

type HttpServerConfig interface {
	HttpAddress() string
	Profiling() bool
}